package razorpay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

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
