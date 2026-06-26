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
)

// TestCreateRefund_ForwardsRefundId verifies that createRefund forwards refund_id and converts amount correctly.
func TestCreateRefund_ForwardsRefundId(t *testing.T) {
	var capturedReq *cf.OrderCreateRefundRequest
	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}

			refundId := "refund_123"
			mockRefund := &cf.RefundEntity{
				RefundId: &refundId,
			}
			jsonData, err := json.Marshal(mockRefund)
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
		HTTPClient:   mockHTTPClient,
	}

	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	req := &domain.CreateRefundRequest{
		RefundID:    "refund_123",
		OrderID:     "order_456",
		AmountMinor: 25000,
		Currency:    domain.Currency("INR"),
	}

	createRefund(context.Background(), adapter, req)

	if capturedReq == nil {
		t.Fatal("request was not captured")
	}
	if capturedReq.RefundId == nil || *capturedReq.RefundId != "refund_123" {
		t.Error("refund_id not forwarded")
	}
	// Verify amount is converted from minor to major (25000 paisa = 250.00 INR)
	if capturedReq.RefundAmount != 250.0 {
		t.Errorf("expected RefundAmount 250.0 (minor=25000, INR), got %v", capturedReq.RefundAmount)
	}
}
