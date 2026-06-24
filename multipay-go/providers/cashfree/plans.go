package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// createPlan creates a new subscription plan on the Cashfree payment gateway.
// Maps the canonical domain.CreatePlanRequest to a Cashfree CreatePlanRequest,
// calls the SDK, and maps the response back to a canonical domain.Plan.
func createPlan(ctx context.Context, a *Adapter, req *domain.CreatePlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Convert amounts from minor to major units
	maxAmountMajor := AmountMinorToMajor(int64(req.MaxAmountMinor), string(req.Currency))
	amountMajor := AmountMinorToMajor(int64(req.AmountMinor), string(req.Currency))

	// Build Cashfree CreatePlanRequest
	// Note: PlanMaxAmount is float32 (required), PlanRecurringAmount is *float32 (optional)
	maxAmount32 := float32(maxAmountMajor)
	amount32 := float32(amountMajor)
	cfReq := &cf.CreatePlanRequest{
		PlanId:              req.PlanID,
		PlanName:            req.PlanName,
		PlanType:            string(req.PlanType),
		PlanMaxAmount:       maxAmount32,
		PlanRecurringAmount: &amount32,
		PlanIntervals:       ptrInt32(req.Interval), // nolint:gosec // Interval is validated as gte=1, safe int32 conversion
		PlanIntervalType:    ptrString(string(req.IntervalType)),
		PlanMaxCycles:       ptrInt32(req.MaxCycles), // nolint:gosec // MaxCycles is validated as gte=0, safe int32 conversion
		PlanNote:            ptrString(req.Note),
	}

	// Call Cashfree SDK
	cfPlan, _, err := a.cfClient.SubsCreatePlanWithContext(
		ctx,
		cfReq,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	if cfPlan == nil {
		return nil, fmt.Errorf("cashfree returned nil plan: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	plan := MapPlanEntityToCanonical(cfPlan)
	return plan, nil
}

// ptrString creates a pointer to a string.
func ptrString(s string) *string {
	return &s
}

// ptrInt32 creates a pointer to an int32.
func ptrInt32(i int32) *int32 {
	return &i
}

// ptrFloat32 creates a pointer to a float32.
func ptrFloat32(f float32) *float32 {
	return &f
}

// getPlan retrieves an existing plan from the Cashfree payment gateway.
// Maps the canonical domain.GetPlanRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.Plan.
func getPlan(ctx context.Context, adapter *Adapter, req *domain.GetPlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.PlanID == "" {
		return nil, fmt.Errorf("PlanID is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK to fetch plan
	cfPlan, _, err := adapter.cfClient.SubsFetchPlanWithContext(
		ctx,
		req.PlanID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 plan not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("plan %s not found: %w", req.PlanID, domain.ErrProviderError)
		}
		return nil, fmt.Errorf("failed to fetch plan from Cashfree: %w", err)
	}

	if cfPlan == nil {
		return nil, fmt.Errorf("cashfree returned nil plan: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	plan := MapPlanEntityToCanonical(cfPlan)
	return plan, nil
}
