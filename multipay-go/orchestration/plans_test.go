package orchestration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/SmrutAI/pedantigo"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// TestCreatePlanRequest_Validation tests required field validation via JSON Unmarshal.
// Required validation can ONLY be tested via Unmarshal, where JSON decoder can distinguish missing from zero.
func TestCreatePlanRequest_Validation(t *testing.T) {
	validator := pedantigo.New[domain.CreatePlanRequest]()

	// Valid JSON with all fields
	validJSON := `{
		"plan_id": "p1",
		"plan_name": "Premium",
		"plan_type": "PERIODIC",
		"max_amount_minor": 100000,
		"currency": "INR",
		"amount_minor": 50000,
		"interval": 1,
		"interval_type": "MONTH"
	}`

	// Case (a): valid JSON unmarshal
	{
		_, err := validator.Unmarshal([]byte(validJSON))
		if err != nil {
			t.Fatalf("expected nil for valid JSON, got %v", err)
		}
	}

	// Case (b): missing plan_name (required field)
	{
		invalidJSON := `{"plan_id":"p1","plan_type":"PERIODIC","max_amount_minor":100000,"currency":"INR","amount_minor":50000,"interval":1,"interval_type":"MONTH"}`
		_, err := validator.Unmarshal([]byte(invalidJSON))
		if err == nil {
			t.Fatalf("expected error for missing plan_name, got nil")
		}
	}

	// Case (c): missing max_amount_minor (required field)
	{
		invalidJSON := `{"plan_id":"p1","plan_name":"Premium","plan_type":"PERIODIC","currency":"INR","amount_minor":50000,"interval":1,"interval_type":"MONTH"}`
		_, err := validator.Unmarshal([]byte(invalidJSON))
		if err == nil {
			t.Fatalf("expected error for missing max_amount_minor, got nil")
		}
	}

	// Case (d): invalid currency (iso4217 validation)
	{
		invalidJSON := `{"plan_id":"p1","plan_name":"Premium","plan_type":"PERIODIC","max_amount_minor":100000,"currency":"XX","amount_minor":50000,"interval":1,"interval_type":"MONTH"}`
		_, err := validator.Unmarshal([]byte(invalidJSON))
		if err == nil {
			t.Fatalf("expected error for invalid currency, got nil")
		}
	}

	// Case (e): missing plan_type (required field)
	{
		invalidJSON := `{"plan_id":"p1","plan_name":"Premium","max_amount_minor":100000,"currency":"INR","amount_minor":50000,"interval":1,"interval_type":"MONTH"}`
		_, err := validator.Unmarshal([]byte(invalidJSON))
		if err == nil {
			t.Fatalf("expected error for missing plan_type, got nil")
		}
	}

	// Case (f): missing plan_id (required field)
	{
		invalidJSON := `{"plan_name":"Premium","plan_type":"PERIODIC","max_amount_minor":100000,"currency":"INR","amount_minor":50000,"interval":1,"interval_type":"MONTH"}`
		_, err := validator.Unmarshal([]byte(invalidJSON))
		if err == nil {
			t.Fatalf("expected error for missing plan_id, got nil")
		}
	}
}

// TestPlanService_CreatePlan_NoCapabilityGate tests the PlanService.CreatePlan method with 3 cases.
func TestPlanService_CreatePlan_NoCapabilityGate(t *testing.T) {
	// Case (a): valid req → returns plan, nil error
	{
		expectedPlan := &domain.Plan{PlanID: "p1"}
		adapter := &fakeAdapter{
			createPlanFunc: func(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error) {
				return expectedPlan, nil
			},
		}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		svc := NewPlanService(domain.ProviderRazorpay, adapter, pipeline, logger, clock)

		req := &domain.CreatePlanRequest{
			PlanID:         "p1",
			PlanName:       "P",
			PlanType:       domain.PlanTypePeriodic,
			MaxAmountMinor: 100000,
			Currency:       "INR",
			AmountMinor:    50000,
			Interval:       1,
			IntervalType:   domain.PlanIntervalMonth,
		}
		plan, err := svc.CreatePlan(context.Background(), req)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if plan != expectedPlan {
			t.Fatalf("expected returned plan to match, got %v", plan)
		}
	}

	// Case (b): adapter returns error → error wraps it
	{
		adapter := &fakeAdapter{
			createPlanFunc: func(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error) {
				return nil, errors.New("boom")
			},
		}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		svc := NewPlanService(domain.ProviderRazorpay, adapter, pipeline, logger, clock)

		req := &domain.CreatePlanRequest{
			PlanID:         "p1",
			PlanName:       "P",
			PlanType:       domain.PlanTypePeriodic,
			MaxAmountMinor: 100000,
			Currency:       "INR",
			AmountMinor:    50000,
			Interval:       1,
			IntervalType:   domain.PlanIntervalMonth,
		}
		_, err := svc.CreatePlan(context.Background(), req)
		if err == nil {
			t.Fatalf("expected error for adapter boom, got nil")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("expected error to contain 'boom', got %v", err)
		}
	}

	// Case (c): nil req → ErrInvalidRequest
	{
		adapter := &fakeAdapter{}
		logger := ports.NewNoopLogger()
		pipeline := hooks.NewPipeline(logger)
		clock := ports.NewRealClock()
		svc := NewPlanService(domain.ProviderRazorpay, adapter, pipeline, logger, clock)

		_, err := svc.CreatePlan(context.Background(), nil)
		if err == nil {
			t.Fatalf("expected error for nil req, got nil")
		}
		if !errors.Is(err, domain.ErrInvalidRequest) {
			t.Fatalf("expected ErrInvalidRequest, got %v", err)
		}
	}
}
