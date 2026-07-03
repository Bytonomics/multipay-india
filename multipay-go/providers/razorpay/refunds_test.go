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
	var capturedBody map[string]any
	mockHTTPClient := &http.Client{
		Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, &capturedBody); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}

			mockResp := map[string]any{
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
		Key:         "rzp_mock_testonly",
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
	notes, ok := capturedBody["notes"].(map[string]any)
	if !ok || len(notes) == 0 {
		t.Error("expected notes to be forwarded from Metadata")
	}
}

// TestListRefunds_UsesPaymentScopedEndpoint verifies that ListRefunds calls Razorpay's
// payment-scoped endpoint (GET /v1/payments/{payment_id}/refunds) using PaymentID, NOT the
// account-wide /v1/refunds endpoint (which would return every refund in the account).
func TestListRefunds_UsesPaymentScopedEndpoint(t *testing.T) {
	var capturedMethod, capturedPath string
	mockHTTPClient := &http.Client{
		Transport: rzRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			mockResp := map[string]any{
				"entity": "collection",
				"count":  1,
				"items": []map[string]any{
					{"id": "refund_1", "payment_id": "pay_123", "status": "processed", "amount": 25000, "currency": "INR"},
				},
			}
			jsonData, marshalErr := json.Marshal(mockResp)
			if marshalErr != nil {
				return nil, marshalErr
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jsonData)))}, nil
		}),
	}

	cfg := &Config{Key: "rzp_mock_testonly", Secret: "test_secret", Environment: domain.EnvironmentSandbox, HTTPClient: mockHTTPClient}
	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	refunds, err := adapter.ListRefunds(context.Background(), &domain.ListRefundsRequest{PaymentID: "pay_123"})
	if err != nil {
		t.Fatalf("ListRefunds returned error: %v", err)
	}

	if capturedMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", capturedMethod)
	}
	if !strings.Contains(capturedPath, "/payments/pay_123/refunds") {
		t.Errorf("expected payment-scoped path /payments/pay_123/refunds, got %s", capturedPath)
	}
	if len(refunds) != 1 {
		t.Errorf("expected 1 refund, got %d", len(refunds))
	}

	// Empty PaymentID must be rejected (the canonical request carries the Razorpay id here).
	if _, emptyErr := adapter.ListRefunds(context.Background(), &domain.ListRefundsRequest{}); emptyErr == nil {
		t.Error("expected error when PaymentID is empty")
	}
}
