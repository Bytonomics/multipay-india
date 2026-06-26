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
