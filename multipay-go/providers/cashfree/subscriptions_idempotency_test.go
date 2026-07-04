package cashfree

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// jsonResp builds an HTTP response with a JSON content type so the Cashfree SDK decodes it
// (an empty content type yields "undefined response type").
func jsonResp(status int, v any) (*http.Response, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(data))),
	}, nil
}

// TestChargeSubscription_SendsIdempotencyKey verifies that the proration charge engages Cashfree's
// native idempotency by sending the caller's PaymentRef as the x-idempotency-key header, so a webhook
// replay returns the original charge instead of debiting the customer twice.
func TestChargeSubscription_SendsIdempotencyKey(t *testing.T) {
	currency := "INR"
	var capturedIdempotencyKey string
	paymentSeen := false

	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Record the native idempotency key from whichever request carries it (the create-payment
			// POST), without letting other requests overwrite it with an empty value. The Cashfree SDK
			// stores this header under the literal lowercase key "x-idempotency-key" (client.go builds
			// req.Header via raw map assignment, not Header.Set, so it is NOT canonicalized) — read the
			// raw map key directly; req.Header.Get canonicalizes the lookup and would miss it.
			if vals := req.Header["x-idempotency-key"]; len(vals) > 0 && vals[0] != "" { //nolint:staticcheck // SA1008: the Cashfree SDK stores this header under the literal non-canonical key; read it verbatim
				capturedIdempotencyKey = vals[0]
			}
			// chargeSubscription first GETs the subscription (to resolve currency), then POSTs the payment.
			if req.Method == http.MethodGet {
				return jsonResp(200, &cf.SubscriptionEntity{PlanDetails: &cf.PlanEntity{PlanCurrency: &currency}})
			}
			paymentSeen = true
			cfPaymentID := "cf_pay_1"
			return jsonResp(200, &cf.CreateSubscriptionPaymentResponse{CfPaymentId: &cfPaymentID})
		}),
	}

	adapter, err := NewAdapter(&Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   mockHTTPClient,
	})
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	_, err = adapter.ChargeSubscription(context.Background(), &domain.ChargeSubscriptionRequest{
		SubscriptionID: "sub_1",
		PaymentRef:     "pay_ref_123",
		AmountMinor:    5000,
		Currency:       "INR",
	})
	if err != nil {
		t.Fatalf("ChargeSubscription returned error: %v", err)
	}
	if !paymentSeen {
		t.Fatal("create-payment request was never made")
	}
	if capturedIdempotencyKey != "pay_ref_123" {
		t.Errorf("x-idempotency-key header = %q, want %q", capturedIdempotencyKey, "pay_ref_123")
	}
}

// TestCancelSubscription_AlreadyCancelled_IsIdempotent verifies that when the provider errors on a
// CANCEL of an already-cancelled subscription, the adapter confirms via a fetch and returns success —
// so a replayed upgrade-finalize can complete instead of looping.
func TestCancelSubscription_AlreadyCancelled_IsIdempotent(t *testing.T) {
	subID := "sub_cxl"
	cancelled := "CANCELLED"

	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// The manage (CANCEL) POST fails; the follow-up fetch (GET) shows the sub is already cancelled.
			if req.Method == http.MethodGet {
				return jsonResp(200, &cf.SubscriptionEntity{SubscriptionId: &subID, SubscriptionStatus: &cancelled})
			}
			return jsonResp(409, map[string]string{"message": "subscription already cancelled"})
		}),
	}

	adapter, err := NewAdapter(&Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
		AccountID:    "test_account",
		Logger:       ports.NewNoopLogger(),
		HTTPClient:   mockHTTPClient,
	})
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	sub, err := adapter.CancelSubscription(context.Background(), &domain.CancelSubscriptionRequest{SubscriptionID: subID})
	if err != nil {
		t.Fatalf("expected idempotent success on already-cancelled subscription, got error: %v", err)
	}
	if sub == nil || sub.Status != domain.SubscriptionStatusCancelled {
		t.Errorf("expected canonical status CANCELLED, got %+v", sub)
	}
}
