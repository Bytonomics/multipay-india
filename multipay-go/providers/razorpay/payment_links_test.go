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

// TestCreatePaymentLink_ForwardsParameters verifies that CreatePaymentLink forwards customer, callback_url, reference_id, and accept_partial to Razorpay.
func TestCreatePaymentLink_ForwardsParameters(t *testing.T) {
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
				"id":     "link_123",
				"status": "created",
				"amount": 50000,
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

	partialPayment := true
	req := &domain.CreatePaymentLinkRequest{
		LinkID:      "link_123",
		AmountMinor: 50000,
		Currency:    domain.Currency("INR"),
		Purpose:     "Payment",
		Customer: &domain.CustomerInfo{
			CustomerID: "cust_456",
			Name:       "Jane Smith",
			Email:      "jane@example.com",
			Phone:      "+919876543210",
		},
		ReturnURL:      "https://example.com/callback",
		PartialPayment: &partialPayment,
	}

	adapter.CreatePaymentLink(context.Background(), req)

	if capturedBody == nil {
		t.Fatal("request was not captured")
	}

	// Assert customer is forwarded
	customer, custOk := capturedBody["customer"].(map[string]interface{})
	if !custOk || len(customer) == 0 {
		t.Error("customer not forwarded")
	} else {
		// Check customer.contact (phone)
		if contact, contactOk := customer["contact"].(string); !contactOk || contact != "+919876543210" {
			t.Errorf("expected customer.contact=%s, got %v", "+919876543210", customer["contact"])
		}
		// Check customer.email
		if email, emailOk := customer["email"].(string); !emailOk || email != "jane@example.com" {
			t.Errorf("expected customer.email=%s, got %v", "jane@example.com", customer["email"])
		}
	}

	// Assert callback_url is forwarded from ReturnURL
	callbackUrl, ok := capturedBody["callback_url"].(string)
	if !ok || callbackUrl != "https://example.com/callback" {
		t.Errorf("expected callback_url=https://example.com/callback, got %v", capturedBody["callback_url"])
	}

	// Assert reference_id is forwarded from LinkID
	referenceId, ok := capturedBody["reference_id"].(string)
	if !ok || referenceId != "link_123" {
		t.Errorf("expected reference_id=link_123, got %v", capturedBody["reference_id"])
	}

	// Assert accept_partial is forwarded
	acceptPartial, ok := capturedBody["accept_partial"].(bool)
	if !ok || !acceptPartial {
		t.Errorf("expected accept_partial=true, got %v", capturedBody["accept_partial"])
	}
}
