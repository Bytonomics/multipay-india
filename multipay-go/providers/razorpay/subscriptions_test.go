package razorpay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

const invoiceFixtureJSON = `{"id":"inv_1","entity":"invoice","payment_id":"pay_1","amount":100,"currency":"INR","status":"paid","order_id":"order_1","paid_at":1481541600,"created_at":1481541534}`

func TestMapInvoiceToSubscriptionPayment(t *testing.T) {
	var invoice razorpayInvoiceResponse
	if err := json.Unmarshal([]byte(invoiceFixtureJSON), &invoice); err != nil {
		t.Fatalf("failed to unmarshal invoice fixture: %v", err)
	}

	// Marshal invoice to bytes for mapper
	invoiceBytes, err := json.Marshal(invoice)
	if err != nil {
		t.Fatalf("failed to marshal invoice: %v", err)
	}

	pmt := mapInvoiceToSubscriptionPayment(&invoice, "sub_1", invoiceBytes)

	if pmt.PaymentID != "pay_1" {
		t.Fatalf("expected PaymentID='pay_1', got '%s'", pmt.PaymentID)
	}

	if pmt.SubscriptionID != "sub_1" {
		t.Fatalf("expected SubscriptionID='sub_1', got '%s'", pmt.SubscriptionID)
	}

	if int64(pmt.AmountMinor) != 100 {
		t.Fatalf("expected AmountMinor=100 (Razorpay native minor, no conversion), got %d", int64(pmt.AmountMinor))
	}

	if pmt.Status != domain.SubPaymentStatusSuccess {
		t.Fatalf("expected Status=SubPaymentStatusSuccess (from 'paid'), got %v", pmt.Status)
	}
}

// TestCreateSubscription_ForwardsCustomerNotify verifies the adapter forwards the caller's
// canonical CustomerNotify to Razorpay's customer_notify (true→1, false→0) and OMITS it when the
// caller leaves it nil — the library imposes NO default.
func TestCreateSubscription_ForwardsCustomerNotify(t *testing.T) {
	newAdapter := func(t *testing.T, capture *map[string]any) *Adapter {
		t.Helper()
		mockHTTPClient := &http.Client{
			Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				if unmarshalErr := json.Unmarshal(body, capture); unmarshalErr != nil {
					t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
				}
				resp := map[string]any{"id": "sub_1", "status": "created", "plan_id": "plan_1"}
				jsonData, merr := json.Marshal(resp)
				if merr != nil {
					return nil, merr
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, nil
			}),
		}
		adapter, err := NewAdapter(&Config{Key: "rzp_mock_testonly", Secret: "test_secret", Environment: domain.EnvironmentSandbox, HTTPClient: mockHTTPClient})
		if err != nil {
			t.Fatalf("failed to create adapter: %v", err)
		}
		return adapter
	}

	baseReq := func() *domain.CreateSubscriptionRequest {
		return &domain.CreateSubscriptionRequest{
			SubscriptionID: "sub_1",
			PlanID:         "plan_1",
			CustomerPhone:  "+919876543210",
			ReturnURL:      "https://example.com/return",
		}
	}

	// false → customer_notify=0 forwarded (JSON numbers decode to float64).
	t.Run("false forwarded as 0", func(t *testing.T) {
		var captured map[string]any
		adapter := newAdapter(t, &captured)
		notify := false
		req := baseReq()
		req.CustomerNotify = &notify
		adapter.CreateSubscription(context.Background(), req)
		if captured == nil {
			t.Fatal("request was not captured")
		}
		v, ok := captured["customer_notify"].(float64)
		if !ok || v != 0 {
			t.Errorf("expected customer_notify=0, got %v", captured["customer_notify"])
		}
	})

	// true → customer_notify=1 forwarded.
	t.Run("true forwarded as 1", func(t *testing.T) {
		var captured map[string]any
		adapter := newAdapter(t, &captured)
		notify := true
		req := baseReq()
		req.CustomerNotify = &notify
		adapter.CreateSubscription(context.Background(), req)
		v, ok := captured["customer_notify"].(float64)
		if !ok || v != 1 {
			t.Errorf("expected customer_notify=1, got %v", captured["customer_notify"])
		}
	})

	// nil → customer_notify omitted (no imposed default).
	t.Run("nil omitted", func(t *testing.T) {
		var captured map[string]any
		adapter := newAdapter(t, &captured)
		adapter.CreateSubscription(context.Background(), baseReq())
		if _, present := captured["customer_notify"]; present {
			t.Errorf("expected customer_notify omitted when nil, got %v", captured["customer_notify"])
		}
	})
}

// TestChargeSubscription_EmptyRemarksGuard verifies that ChargeSubscription
// returns an error before making the API call when Remarks is empty.
func TestChargeSubscription_EmptyRemarksGuard(t *testing.T) {
	apiCalled := false
	mockHTTPClient := &http.Client{
		Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			apiCalled = true
			resp := map[string]any{"error": "should not be called"}
			jsonData, err := json.Marshal(resp)
			if err != nil {
				return nil, err
			}
			return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, err
		}),
	}
	adapter, err := NewAdapter(&Config{
		Key:         "rzp_mock_testonly",
		Secret:      "test_secret",
		Environment: domain.EnvironmentSandbox,
		HTTPClient:  mockHTTPClient,
	})
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	req := &domain.ChargeSubscriptionRequest{
		SubscriptionID: "sub_123",
		PaymentRef:     "payment_456",
		AmountMinor:    50000,
		Currency:       "INR",
		Remarks:        "", // Empty remarks — should be rejected
	}

	_, chargeErr := adapter.ChargeSubscription(context.Background(), req)
	if chargeErr == nil {
		t.Errorf("expected error for empty Remarks, got nil")
	}

	if apiCalled {
		t.Errorf("API should NOT have been called when Remarks is empty")
	}
}

// TestCreateSubscription_FirstChargeWithMandate verifies that when FirstChargeWithMandate
// is true, the adapter appends a first-period addon with the plan's amount+currency and
// sets start_at to FirstChargeTime + interval.
func TestCreateSubscription_FirstChargeWithMandate(t *testing.T) {
	fixed := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)

	// Test A: monthly schedule — addon carries plan amount+currency, start_at = now+1 month
	t.Run("monthly: addon carries plan amount+currency, start_at = now+1 month", func(t *testing.T) {
		var captured map[string]any
		mockHTTPClient := &http.Client{
			Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				if unmarshalErr := json.Unmarshal(body, &captured); unmarshalErr != nil {
					t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
				}
				resp := map[string]any{"id": "sub_1", "status": "created", "plan_id": "plan_x"}
				jsonData, merr := json.Marshal(resp)
				if merr != nil {
					return nil, merr
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, nil
			}),
		}
		adapter, err := NewAdapter(&Config{
			Key:         "rzp_mock_testonly",
			Secret:      "test_secret",
			Environment: domain.EnvironmentSandbox,
			HTTPClient:  mockHTTPClient,
		})
		if err != nil {
			t.Fatalf("failed to create adapter: %v", err)
		}

		req := &domain.CreateSubscriptionRequest{
			PlanID:                 "plan_x",
			ReturnURL:              "https://example.com/return",
			CustomerPhone:          "9876543210",
			CustomerEmail:          "test@example.com",
			FirstChargeWithMandate: true,
			FirstChargeTime:        &fixed,
			RecurringAmountMinor:   49900,
			RecurringInterval:      1,
			RecurringIntervalType:  domain.PlanIntervalMonth,
			RecurringCurrency:      domain.Currency("INR"),
		}

		_, createErr := adapter.CreateSubscription(context.Background(), req)
		if createErr != nil {
			t.Fatalf("expected no error, got %v", createErr)
		}

		if captured == nil {
			t.Fatal("request was not captured")
		}

		// Check addons
		addonsRaw, ok := captured["addons"].([]any)
		if !ok {
			t.Errorf("expected addons to be an array, got %T", captured["addons"])
		}
		if len(addonsRaw) < 1 {
			t.Fatalf("expected at least 1 addon, got %d", len(addonsRaw))
		}

		// Get the LAST addon (the first-charge addon)
		lastAddonRaw := addonsRaw[len(addonsRaw)-1]
		lastAddonMap, ok := lastAddonRaw.(map[string]any)
		if !ok {
			t.Fatalf("expected addon to be a map, got %T", lastAddonRaw)
		}

		itemRaw, ok := lastAddonMap["item"].(map[string]any)
		if !ok {
			t.Fatalf("expected addon.item to be a map, got %T", lastAddonMap["item"])
		}

		amount, ok := itemRaw["amount"].(float64)
		if !ok || amount != 49900 {
			t.Errorf("expected addon.item.amount == 49900, got %v", itemRaw["amount"])
		}

		currency, ok := itemRaw["currency"].(string)
		if !ok || currency != "INR" {
			t.Errorf("expected addon.item.currency == 'INR', got %q", itemRaw["currency"])
		}

		name, ok := itemRaw["name"].(string)
		if !ok || name != "First billing period" {
			t.Errorf("expected addon.item.name == 'First billing period', got %q", itemRaw["name"])
		}

		// Check start_at
		startAtRaw, ok := captured["start_at"].(float64)
		if !ok {
			t.Fatalf("expected start_at to be a float64, got %T", captured["start_at"])
		}
		startAt := int64(startAtRaw)
		expectedStartAt := fixed.AddDate(0, 1, 0).Unix()
		if startAt != expectedStartAt {
			t.Errorf("expected start_at == %d, got %d", expectedStartAt, startAt)
		}
	})

	// Test B: interval offsets (table) — day, week, year with varying amounts
	t.Run("interval offsets (table)", func(t *testing.T) {
		tests := []struct {
			name            string
			intervalType    domain.PlanIntervalType
			interval        int32
			currency        domain.Currency
			amount          domain.AmountMinor
			expectedStartAt time.Time
		}{
			{
				name:            "Day, interval 10",
				intervalType:    domain.PlanIntervalDay,
				interval:        10,
				currency:        domain.Currency("INR"),
				amount:          100,
				expectedStartAt: fixed.AddDate(0, 0, 10),
			},
			{
				name:            "Week, interval 2",
				intervalType:    domain.PlanIntervalWeek,
				interval:        2,
				currency:        domain.Currency("INR"),
				amount:          100,
				expectedStartAt: fixed.AddDate(0, 0, 14),
			},
			{
				name:            "Year, interval 1",
				intervalType:    domain.PlanIntervalYear,
				interval:        1,
				currency:        domain.Currency("INR"),
				amount:          100,
				expectedStartAt: fixed.AddDate(1, 0, 0),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var captured map[string]any
				mockHTTPClient := &http.Client{
					Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
						body, err := io.ReadAll(req.Body)
						if err != nil {
							return nil, err
						}
						if unmarshalErr := json.Unmarshal(body, &captured); unmarshalErr != nil {
							t.Fatalf("failed to unmarshal request body: %v", unmarshalErr)
						}
						resp := map[string]any{"id": "sub_1", "status": "created", "plan_id": "plan_x"}
						jsonData, merr := json.Marshal(resp)
						if merr != nil {
							return nil, merr
						}
						return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, nil
					}),
				}
				adapter, err := NewAdapter(&Config{
					Key:         "rzp_mock_testonly",
					Secret:      "test_secret",
					Environment: domain.EnvironmentSandbox,
					HTTPClient:  mockHTTPClient,
				})
				if err != nil {
					t.Fatalf("failed to create adapter: %v", err)
				}

				req := &domain.CreateSubscriptionRequest{
					PlanID:                 "plan_x",
					ReturnURL:              "https://example.com/return",
					CustomerPhone:          "9876543210",
					CustomerEmail:          "test@example.com",
					FirstChargeWithMandate: true,
					FirstChargeTime:        &fixed,
					RecurringAmountMinor:   tt.amount,
					RecurringInterval:      tt.interval,
					RecurringIntervalType:  tt.intervalType,
					RecurringCurrency:      tt.currency,
				}

				_, createErr := adapter.CreateSubscription(context.Background(), req)
				if createErr != nil {
					t.Fatalf("expected no error, got %v", createErr)
				}

				if captured == nil {
					t.Fatal("request was not captured")
				}

				// Check start_at
				startAtRaw, ok := captured["start_at"].(float64)
				if !ok {
					t.Fatalf("expected start_at to be a float64, got %T", captured["start_at"])
				}
				startAt := int64(startAtRaw)
				expectedStartAt := tt.expectedStartAt.Unix()
				if startAt != expectedStartAt {
					t.Errorf("expected start_at == %d, got %d", expectedStartAt, startAt)
				}

				// Check last addon's currency and amount
				addonsRaw, ok := captured["addons"].([]any)
				if !ok || len(addonsRaw) < 1 {
					t.Fatalf("expected at least 1 addon")
				}
				lastAddonMap, ok := addonsRaw[len(addonsRaw)-1].(map[string]any)
				if !ok {
					t.Fatalf("expected addon to be a map")
				}
				itemRaw, ok := lastAddonMap["item"].(map[string]any)
				if !ok {
					t.Fatalf("expected addon.item to be a map")
				}

				amount, ok := itemRaw["amount"].(float64)
				if !ok || amount != float64(tt.amount) {
					t.Errorf("expected addon.item.amount == %d, got %v", tt.amount, itemRaw["amount"])
				}

				currency, ok := itemRaw["currency"].(string)
				if !ok || currency != string(tt.currency) {
					t.Errorf("expected addon.item.currency == %q, got %q", string(tt.currency), itemRaw["currency"])
				}
			})
		}
	})

	// Test C: missing schedule fields for existing plan_id → deterministic error, no HTTP call
	t.Run("missing schedule fields for existing plan_id → deterministic error, no HTTP call", func(t *testing.T) {
		apiCalled := false
		mockHTTPClient := &http.Client{
			Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				apiCalled = true
				t.Fatalf("no HTTP call expected")
				return &http.Response{}, errors.New("mock error")
			}),
		}
		adapter, err := NewAdapter(&Config{
			Key:         "rzp_mock_testonly",
			Secret:      "test_secret",
			Environment: domain.EnvironmentSandbox,
			HTTPClient:  mockHTTPClient,
		})
		if err != nil {
			t.Fatalf("failed to create adapter: %v", err)
		}

		// Leave RecurringAmountMinor, RecurringInterval, RecurringIntervalType, RecurringCurrency at zero-values
		req := &domain.CreateSubscriptionRequest{
			PlanID:                 "plan_x",
			ReturnURL:              "https://example.com/return",
			CustomerPhone:          "9876543210",
			CustomerEmail:          "test@example.com",
			FirstChargeWithMandate: true,
			FirstChargeTime:        &fixed,
			PlanDetails:            nil, // no inline plan
			// RecurringAmountMinor, RecurringInterval, etc. left at zero-values
		}

		_, createErr := adapter.CreateSubscription(context.Background(), req)
		if createErr == nil {
			t.Errorf("expected error for missing schedule fields, got nil")
		}

		if !errors.Is(createErr, domain.ErrInvalidRequest) {
			t.Errorf("expected error to be ErrInvalidRequest, got %v", createErr)
		}

		if apiCalled {
			t.Errorf("API should NOT have been called for missing schedule fields")
		}
	})

	// Test D: missing currency alone → deterministic error
	t.Run("missing currency alone → deterministic error", func(t *testing.T) {
		apiCalled := false
		mockHTTPClient := &http.Client{
			Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				apiCalled = true
				t.Fatalf("no HTTP call expected")
				return &http.Response{}, errors.New("mock error")
			}),
		}
		adapter, err := NewAdapter(&Config{
			Key:         "rzp_mock_testonly",
			Secret:      "test_secret",
			Environment: domain.EnvironmentSandbox,
			HTTPClient:  mockHTTPClient,
		})
		if err != nil {
			t.Fatalf("failed to create adapter: %v", err)
		}

		req := &domain.CreateSubscriptionRequest{
			PlanID:                 "plan_x",
			ReturnURL:              "https://example.com/return",
			CustomerPhone:          "9876543210",
			CustomerEmail:          "test@example.com",
			FirstChargeWithMandate: true,
			FirstChargeTime:        &fixed,
			PlanDetails:            nil,
			RecurringAmountMinor:   49900,
			RecurringInterval:      1,
			RecurringIntervalType:  domain.PlanIntervalMonth,
			RecurringCurrency:      "", // missing currency
		}

		_, createErr := adapter.CreateSubscription(context.Background(), req)
		if createErr == nil {
			t.Errorf("expected error for missing currency, got nil")
		}

		if !errors.Is(createErr, domain.ErrInvalidRequest) {
			t.Errorf("expected error to be ErrInvalidRequest, got %v", createErr)
		}

		if apiCalled {
			t.Errorf("API should NOT have been called for missing currency")
		}
	})
}
