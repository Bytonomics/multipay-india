package cashfree

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// TestBuildInlinePlanDetails_AllFields verifies that buildInlinePlanDetails correctly maps all inline plan fields,
// including amount conversion from minor to major units.
func TestBuildInlinePlanDetails_AllFields(t *testing.T) {
	req := &domain.CreatePlanRequest{
		PlanID:         "plan_123",
		PlanName:       "Premium Plan",
		PlanType:       domain.PlanTypePeriodic,
		Currency:       "INR",
		AmountMinor:    50000,  // ₹500.00
		MaxAmountMinor: 100000, // ₹1000.00
		Interval:       1,
		IntervalType:   domain.PlanIntervalMonth,
	}

	result := buildInlinePlanDetails(req)

	// Assert PlanId
	if result.PlanId == nil || *result.PlanId != "plan_123" {
		t.Errorf("expected PlanId=plan_123, got %v", result.PlanId)
	}

	// Assert PlanName
	if result.PlanName == nil || *result.PlanName != "Premium Plan" {
		t.Errorf("expected PlanName=Premium Plan, got %v", result.PlanName)
	}

	// Assert PlanType
	if result.PlanType == nil || *result.PlanType != "PERIODIC" {
		t.Errorf("expected PlanType=PERIODIC, got %v", result.PlanType)
	}

	// Assert PlanAmount (minor to major conversion: 50000 paisa = 500.00 INR)
	if result.PlanAmount == nil || *result.PlanAmount != 500.0 {
		t.Errorf("expected PlanAmount=500.0 (minor=50000, INR), got %v", result.PlanAmount)
	}

	// Assert PlanMaxAmount (minor to major conversion: 100000 paisa = 1000.00 INR)
	if result.PlanMaxAmount == nil || *result.PlanMaxAmount != 1000.0 {
		t.Errorf("expected PlanMaxAmount=1000.0 (minor=100000, INR), got %v", result.PlanMaxAmount)
	}

	// Assert PlanIntervals
	if result.PlanIntervals == nil || *result.PlanIntervals != 1 {
		t.Errorf("expected PlanIntervals=1, got %v", result.PlanIntervals)
	}

	// Assert PlanIntervalType
	if result.PlanIntervalType == nil {
		t.Error("expected PlanIntervalType to be non-nil")
	} else if *result.PlanIntervalType != "MONTH" {
		t.Errorf("expected PlanIntervalType=MONTH, got %v", *result.PlanIntervalType)
	}
}

// TestCreateSubscription_ForwardsCustomerName verifies that createSubscription forwards customer_name, return_url, and tags to Cashfree.
func TestCreateSubscription_ForwardsCustomerName(t *testing.T) {
	var capturedReq *cf.CreateSubscriptionRequest
	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}

			subId := "cf_sub_123"
			mockSub := &cf.SubscriptionEntity{
				SubscriptionId: &subId,
			}
			jsonData, err := json.Marshal(mockSub)
			if err != nil {
				return nil, err
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(jsonData))),
			}, nil
		}),
	}

	cfg := &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   mockHTTPClient,
	}

	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	req := &domain.CreateSubscriptionRequest{
		SubscriptionID: "sub_123",
		PlanID:         "plan_456",
		CustomerEmail:  "test@example.com",
		CustomerPhone:  "9876543210",
		CustomerName:   "John Doe",
		ReturnURL:      "https://example.com/mandate",
		Tags: map[string]string{
			"type":    "premium",
			"channel": "app",
		},
	}

	createSubscription(context.Background(), adapter, req)

	if capturedReq == nil {
		t.Fatal("request was not captured")
	}

	// Assert CustomerDetails.CustomerName is forwarded
	if capturedReq.CustomerDetails.CustomerName == nil || *capturedReq.CustomerDetails.CustomerName != "John Doe" {
		t.Error("CustomerDetails.CustomerName not forwarded")
	}

	// Assert SubscriptionMeta.ReturnUrl is forwarded
	if capturedReq.SubscriptionMeta == nil || capturedReq.SubscriptionMeta.ReturnUrl == nil || *capturedReq.SubscriptionMeta.ReturnUrl != "https://example.com/mandate" {
		t.Error("SubscriptionMeta.ReturnUrl not forwarded")
	}

	// Assert SubscriptionTags are forwarded
	if capturedReq.SubscriptionTags == nil || len(capturedReq.SubscriptionTags) == 0 {
		t.Error("SubscriptionTags not forwarded")
	}
}

// newCfSubCaptureAdapter returns a Cashfree adapter whose transport captures the outbound JSON
// body into *captured and replies with the given typed SDK entity (via jsonResp so the SDK
// decodes it — an empty content type yields "undefined response type").
func newCfSubCaptureAdapter(t *testing.T, captured any, respEntity any) *Adapter {
	t.Helper()
	client := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, captured); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}
			return jsonResp(200, respEntity)
		}),
	}
	cfg := &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   client,
	}
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	return adapter
}

// TestCreateSubscription_ForwardsAuthAndMetaAndSplitsAndBank verifies that the optional Cashfree
// subscription fields (authorization_details, subscription_meta extras, TPV bank details, and
// Easy-Split payment splits) reach the Cashfree request body.
func TestCreateSubscription_ForwardsAuthAndMetaAndSplitsAndBank(t *testing.T) {
	var capturedReq cf.CreateSubscriptionRequest
	subID := "cf_sub_1"
	adapter := newCfSubCaptureAdapter(t, &capturedReq, &cf.SubscriptionEntity{SubscriptionId: &subID})

	refund := true
	req := &domain.CreateSubscriptionRequest{
		SubscriptionID: "sub_1",
		PlanDetails: &domain.CreatePlanRequest{
			PlanID:         "plan_1",
			PlanName:       "Plan 1",
			PlanType:       domain.PlanTypePeriodic,
			Currency:       "INR",
			AmountMinor:    50000,
			MaxAmountMinor: 100000,
			Interval:       1,
			IntervalType:   domain.PlanIntervalMonth,
		},
		CustomerEmail: "test@example.com",
		CustomerPhone: "9876543210",
		ReturnURL:     "https://example.com/mandate",
		AuthorizationDetails: &domain.SubscriptionAuthorizationDetails{
			AuthorizationAmountMinor:  10000, // ₹100.00
			AuthorizationAmountRefund: &refund,
			PaymentMethods:            []string{"upi", "enach"},
		},
		Meta: &domain.SubscriptionMeta{
			NotificationChannel: []string{"EMAIL", "SMS"},
			SessionIDExpiry:     "2030-01-01T00:00:00+05:30",
		},
		BankDetails: &domain.SubscriptionBankDetails{
			AccountHolderName: "John Doe",
			AccountNumber:     "1234567890",
			IFSC:              "HDFC0000001",
			BankCode:          "3001",
			AccountType:       "SAVINGS",
		},
		PaymentSplits: []domain.SubscriptionPaymentSplit{
			{VendorID: "vendor_1", Percentage: 30},
		},
	}

	if _, err := createSubscription(context.Background(), adapter, req); err != nil {
		t.Fatalf("createSubscription returned error: %v", err)
	}

	// authorization_details
	if capturedReq.AuthorizationDetails == nil {
		t.Fatal("authorization_details not forwarded")
	}
	if capturedReq.AuthorizationDetails.AuthorizationAmount == nil || *capturedReq.AuthorizationDetails.AuthorizationAmount != 100.0 {
		t.Errorf("authorization_amount not forwarded/converted (want 100.0): %v", capturedReq.AuthorizationDetails.AuthorizationAmount)
	}
	if capturedReq.AuthorizationDetails.AuthorizationAmountRefund == nil || !*capturedReq.AuthorizationDetails.AuthorizationAmountRefund {
		t.Error("authorization_amount_refund not forwarded")
	}
	if len(capturedReq.AuthorizationDetails.PaymentMethods) != 2 {
		t.Errorf("payment_methods not forwarded: %v", capturedReq.AuthorizationDetails.PaymentMethods)
	}

	// subscription_meta extras
	if capturedReq.SubscriptionMeta == nil {
		t.Fatal("subscription_meta not forwarded")
	}
	if len(capturedReq.SubscriptionMeta.NotificationChannel) != 2 {
		t.Errorf("notification_channel not forwarded: %v", capturedReq.SubscriptionMeta.NotificationChannel)
	}
	if capturedReq.SubscriptionMeta.SessionIdExpiry == nil || *capturedReq.SubscriptionMeta.SessionIdExpiry != "2030-01-01T00:00:00+05:30" {
		t.Errorf("session_id_expiry not forwarded: %v", capturedReq.SubscriptionMeta.SessionIdExpiry)
	}
	if capturedReq.SubscriptionMeta.ReturnUrl == nil || *capturedReq.SubscriptionMeta.ReturnUrl != "https://example.com/mandate" {
		t.Error("return_url not preserved alongside meta extras")
	}

	// TPV bank details
	if capturedReq.CustomerDetails.CustomerBankAccountNumber == nil || *capturedReq.CustomerDetails.CustomerBankAccountNumber != "1234567890" {
		t.Error("customer_bank_account_number not forwarded")
	}
	if capturedReq.CustomerDetails.CustomerBankIfsc == nil || *capturedReq.CustomerDetails.CustomerBankIfsc != "HDFC0000001" {
		t.Error("customer_bank_ifsc not forwarded")
	}
	if capturedReq.CustomerDetails.CustomerBankAccountType == nil || *capturedReq.CustomerDetails.CustomerBankAccountType != "SAVINGS" {
		t.Error("customer_bank_account_type not forwarded")
	}

	// Easy Split
	if len(capturedReq.SubscriptionPaymentSplits) != 1 {
		t.Fatalf("subscription_payment_splits not forwarded: %v", capturedReq.SubscriptionPaymentSplits)
	}
	if capturedReq.SubscriptionPaymentSplits[0].VendorId == nil || *capturedReq.SubscriptionPaymentSplits[0].VendorId != "vendor_1" {
		t.Error("split vendor_id not forwarded")
	}
	if capturedReq.SubscriptionPaymentSplits[0].Percentage == nil || *capturedReq.SubscriptionPaymentSplits[0].Percentage != 30 {
		t.Error("split percentage not forwarded")
	}
}

// TestChargeSubscription_ForwardsPaymentScheduleDate verifies the Cashfree future-dated charge
// field (payment_schedule_date, date-only) reaches the SDK request.
func TestChargeSubscription_ForwardsPaymentScheduleDate(t *testing.T) {
	var capturedReq cf.CreateSubscriptionPaymentRequest
	currency := "INR"
	cfPaymentID := "cf_pay_1"
	client := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// The charge path first GETs the subscription (for currency), then POSTs the payment.
			// Only capture the payment POST body (it carries payment_schedule_date).
			if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/pay") {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
					t.Fatalf("failed to unmarshal payment body: %v (body: %s)", unmarshalErr, string(body))
				}
				return jsonResp(200, &cf.CreateSubscriptionPaymentResponse{CfPaymentId: &cfPaymentID})
			}
			// Fetch-subscription response: supply currency via plan_details.
			return jsonResp(200, &cf.SubscriptionEntity{PlanDetails: &cf.PlanEntity{PlanCurrency: &currency}})
		}),
	}
	cfg := &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   client,
	}
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	sched := time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)
	req := &domain.ChargeSubscriptionRequest{
		SubscriptionID:      "sub_1",
		PaymentRef:          "charge_1",
		AmountMinor:         20000,
		Currency:            "INR",
		PaymentScheduleDate: &sched,
	}

	if _, err := chargeSubscription(context.Background(), adapter, req); err != nil {
		t.Fatalf("chargeSubscription returned error: %v", err)
	}
	if capturedReq.PaymentScheduleDate == nil || *capturedReq.PaymentScheduleDate != "2030-06-15" {
		t.Errorf("payment_schedule_date not forwarded (want 2030-06-15): %v", capturedReq.PaymentScheduleDate)
	}
}

// TestResumeSubscription_SendsNextScheduledTime verifies the ★ correctness fix: the ACTIVATE
// action forwards action_details.next_scheduled_time (Cashfree rejects ACTIVATE without it).
func TestResumeSubscription_SendsNextScheduledTime(t *testing.T) {
	var capturedReq cf.ManageSubscriptionRequest
	subID := "sub_1"
	status := "ACTIVE"
	adapter := newCfSubCaptureAdapter(t, &capturedReq, &cf.SubscriptionEntity{SubscriptionId: &subID, SubscriptionStatus: &status})

	next := time.Date(2030, 3, 1, 10, 0, 0, 0, time.UTC)
	req := &domain.ResumeSubscriptionRequest{
		SubscriptionID:    "sub_1",
		NextScheduledTime: &next,
	}

	if _, err := resumeSubscription(context.Background(), adapter, req); err != nil {
		t.Fatalf("resumeSubscription returned error: %v", err)
	}
	if capturedReq.Action != "ACTIVATE" {
		t.Errorf("action = %q, want ACTIVATE", capturedReq.Action)
	}
	if capturedReq.ActionDetails == nil || capturedReq.ActionDetails.NextScheduledTime == nil {
		t.Fatal("action_details.next_scheduled_time not forwarded on ACTIVATE")
	}
	if *capturedReq.ActionDetails.NextScheduledTime != "2030-03-01T10:00:00+00:00" {
		t.Errorf("next_scheduled_time = %q, want 2030-03-01T10:00:00+00:00", *capturedReq.ActionDetails.NextScheduledTime)
	}
}

// TestMapSubscriptionEntityToCanonical_AuthLinkFromSessionID verifies that MapSubscriptionEntityToCanonical
// populates domain.Subscription.AuthLink from Cashfree SubscriptionEntity.SubscriptionSessionId.
func TestMapSubscriptionEntityToCanonical_AuthLinkFromSessionID(t *testing.T) {
	subID := "sub_merchant_1"
	cfSubID := "cf_sub_1"
	status := "INITIALIZED"
	sessionID := "sub_session_abc123"

	entity := &cf.SubscriptionEntity{
		SubscriptionId:        &subID,
		CfSubscriptionId:      &cfSubID,
		SubscriptionStatus:    &status,
		SubscriptionSessionId: &sessionID,
	}

	got, err := MapSubscriptionEntityToCanonical(entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil subscription")
	}
	if got.AuthLink != sessionID {
		t.Errorf("AuthLink = %q, want %q (must come from SubscriptionSessionId)", got.AuthLink, sessionID)
	}
	if got.ProviderSubscriptionID != cfSubID {
		t.Errorf("ProviderSubscriptionID = %q, want %q", got.ProviderSubscriptionID, cfSubID)
	}
}

// TestCreateSubscription_ForwardsCfOrderID verifies the adapter forwards canonical CfOrderID to
// Cashfree's cf_order_id, and omits it when empty (the library imposes no default).
func TestCreateSubscription_ForwardsCfOrderID(t *testing.T) {
	newAdapter := func(t *testing.T, capture **cf.CreateSubscriptionRequest) *Adapter {
		t.Helper()
		mockHTTPClient := &http.Client{
			Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				if unmarshalErr := json.Unmarshal(body, capture); unmarshalErr != nil {
					t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
				}
				subID := "cf_sub_1"
				mockSub := &cf.SubscriptionEntity{SubscriptionId: &subID}
				jsonData, merr := json.Marshal(mockSub)
				if merr != nil {
					return nil, merr
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, nil
			}),
		}
		adapter, err := NewAdapter(&Config{ClientID: "test_client_id", ClientSecret: "test_client_secret", Environment: domain.EnvironmentSandbox, AccountID: "test_account", Logger: ports.NewNoopLogger(), HTTPClient: mockHTTPClient})
		if err != nil {
			t.Fatalf("failed to create adapter: %v", err)
		}
		return adapter
	}

	baseReq := func() *domain.CreateSubscriptionRequest {
		return &domain.CreateSubscriptionRequest{
			SubscriptionID: "sub_1",
			PlanID:         "plan_1",
			CustomerEmail:  "test@example.com",
			CustomerPhone:  "9876543210",
			ReturnURL:      "https://example.com/mandate",
		}
	}

	// CfOrderID set → cf_order_id forwarded.
	t.Run("set forwarded", func(t *testing.T) {
		var captured *cf.CreateSubscriptionRequest
		adapter := newAdapter(t, &captured)
		req := baseReq()
		req.CfOrderID = "cforder_123"
		createSubscription(context.Background(), adapter, req)
		if captured == nil {
			t.Fatal("request was not captured")
		}
		if captured.CfOrderId == nil || *captured.CfOrderId != "cforder_123" {
			t.Errorf("expected cf_order_id=cforder_123, got %v", captured.CfOrderId)
		}
	})

	// CfOrderID empty → cf_order_id omitted.
	t.Run("empty omitted", func(t *testing.T) {
		var captured *cf.CreateSubscriptionRequest
		adapter := newAdapter(t, &captured)
		createSubscription(context.Background(), adapter, baseReq())
		if captured == nil {
			t.Fatal("request was not captured")
		}
		if captured.CfOrderId != nil {
			t.Errorf("expected cf_order_id nil when empty, got %v", *captured.CfOrderId)
		}
	})
}
