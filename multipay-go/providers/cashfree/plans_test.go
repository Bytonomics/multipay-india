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

// TestCreatePlan_SendsPlanCurrency is the regression guard for the drift where the standalone
// createPlan mapper omitted plan_currency (Cashfree requires it and rejects a blank value). It
// captures the outbound SDK request body and asserts plan_currency is present and equals the
// canonical request's Currency.
func TestCreatePlan_SendsPlanCurrency(t *testing.T) {
	var capturedReq *cf.CreatePlanRequest
	mockHTTPClient := &http.Client{
		Transport: cfRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if unmarshalErr := json.Unmarshal(body, &capturedReq); unmarshalErr != nil {
				t.Fatalf("failed to unmarshal request body: %v (body: %s)", unmarshalErr, string(body))
			}

			planID := "plan_123"
			mockPlan := &cf.PlanEntity{PlanId: &planID}
			jsonData, err := json.Marshal(mockPlan)
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

	req := &domain.CreatePlanRequest{
		PlanID:         "plan_123",
		PlanName:       "Pro MONTHLY IN",
		PlanType:       domain.PlanTypePeriodic,
		Currency:       "INR",
		AmountMinor:    50000,  // ₹500.00
		MaxAmountMinor: 100000, // ₹1000.00
		Interval:       1,
		IntervalType:   domain.PlanIntervalMonth,
	}

	// The return value is intentionally ignored — the assertion is on the captured outbound request.
	createPlan(context.Background(), adapter, req)

	if capturedReq == nil {
		t.Fatal("request was not captured")
	}
	if capturedReq.PlanCurrency == nil {
		t.Fatal("plan_currency must be sent to Cashfree, got nil (omitted from request)")
	}
	if *capturedReq.PlanCurrency != "INR" {
		t.Errorf("plan_currency = %q, want %q", *capturedReq.PlanCurrency, "INR")
	}
}
