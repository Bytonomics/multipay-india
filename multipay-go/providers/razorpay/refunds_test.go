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

// TestCreateRefund_ForwardsPaymentID verifies that CreateRefund uses PaymentID (not OrderID) as the Razorpay payment_id,
// and forwards RefundID as receipt and Metadata as notes.
func TestCreateRefund_ForwardsPaymentID(t *testing.T) {
	var capturedBody map[string]interface{}
	mockHTTPClient := &http.Client{
		Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, &capturedBody); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}

			mockResp := map[string]interface{}{
				"id":       "refund_123",
				"status":   "processed",
				"amount":   250,
				"currency": "INR",
			}
			jsonData, err := json.Marshal(mockResp)
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
		Key:         "rzp_test_abc123",
		Secret:      "test_secret",
		Environment: domain.EnvironmentSandbox,
		HTTPClient:  mockHTTPClient,
	}

	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	req := &domain.CreateRefundRequest{
		PaymentID:   "pay_123",
		OrderID:     "order_456",
		RefundID:    "refund_123",
		AmountMinor: 25000,
		Currency:    domain.Currency("INR"),
		Metadata: domain.Metadata{
			"reason": "customer_request",
		},
	}

	adapter.CreateRefund(context.Background(), req)

	if capturedBody == nil {
		t.Fatal("request was not captured")
	}

	// Assert payment_id is set from PaymentID (NOT from OrderID)
	paymentId, ok := capturedBody["payment_id"].(string)
	if !ok || paymentId != "pay_123" {
		t.Errorf("expected payment_id=pay_123 (from PaymentID), got %v", capturedBody["payment_id"])
	}

	// Verify OrderID is NOT used as payment_id
	if capturedBody["payment_id"] == "order_456" {
		t.Error("payment_id must come from PaymentID, not OrderID")
	}

	// Assert receipt (RefundID) is forwarded
	receipt, ok := capturedBody["receipt"].(string)
	if !ok || receipt != "refund_123" {
		t.Errorf("expected receipt=refund_123 (from RefundID), got %v", capturedBody["receipt"])
	}

	// Assert notes are forwarded from Metadata
	notes, ok := capturedBody["notes"].(map[string]interface{})
	if !ok || len(notes) == 0 {
		t.Error("expected notes to be forwarded from Metadata")
	}
}
