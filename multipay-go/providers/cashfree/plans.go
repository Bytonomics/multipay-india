package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

// createPlan creates a new subscription plan on the Cashfree payment gateway.
// Maps the canonical domain.CreatePlanRequest to a Cashfree CreatePlanRequest,
// calls the SDK, and maps the response back to a canonical domain.Plan.
func createPlan(ctx context.Context, a *Adapter, req *domain.CreatePlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Convert amounts from minor to major units
	maxAmountMajor := currencyutils.AmountMinorToMajor(int64(req.MaxAmountMinor), string(req.Currency))
	amountMajor := currencyutils.AmountMinorToMajor(int64(req.AmountMinor), string(req.Currency))

	// Build Cashfree CreatePlanRequest.
	// Mandatory fields (validated non-empty at the boundary by CreatePlanRequest.Validate) are
	// set directly. Optional fields use the nil-if-empty helpers so omitempty drops them.
	maxAmount32 := float32(maxAmountMajor)
	cfReq := &cf.CreatePlanRequest{
		PlanId:        req.PlanID,
		PlanName:      req.PlanName,
		PlanType:      string(req.PlanType),
		PlanCurrency:  ptrString(string(req.Currency)),
		PlanMaxAmount: maxAmount32,
		PlanMaxCycles: optInt32(req.MaxCycles),
		PlanNote:      optStr(req.Note),
	}

	// Conditional (PERIODIC only): the per-cycle recurring amount and the interval are
	// meaningless for ON_DEMAND plans and must be omitted, not sent as zero/empty.
	if req.PlanType == domain.PlanTypePeriodic {
		amount32 := float32(amountMajor)
		cfReq.PlanRecurringAmount = &amount32
		cfReq.PlanIntervals = optInt32(req.Interval)
		cfReq.PlanIntervalType = optStr(string(req.IntervalType))
	}

	// Call Cashfree SDK
	cfPlan, httpResp, err := a.cfClient.SubsCreatePlanWithContext(
		ctx,
		cfReq,
		nil, // xRequestId
		nil, // xIdempotencyKey
		a.httpClient,
	)
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			_ = httpResp.Body.Close()
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	if cfPlan == nil {
		return nil, fmt.Errorf("cashfree returned nil plan: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	plan, err := MapPlanEntityToCanonical(cfPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to map plan: %w", err)
	}
	return plan, nil
}

// ptrString creates a pointer to a string.
func ptrString(s string) *string {
	return &s
}

// optStr returns nil for an empty string so that an optional `,omitempty` field is dropped
// rather than serialized as "". A non-nil pointer to "" is NOT dropped by omitempty.
func optStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// optInt32 returns nil for a zero value so that an optional `,omitempty` field is dropped
// rather than serialized as 0. A non-nil pointer to 0 is NOT dropped by omitempty.
func optInt32(i int32) *int32 {
	if i == 0 {
		return nil
	}
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
	cfPlan, httpResp, err := adapter.cfClient.SubsFetchPlanWithContext(
		ctx,
		req.PlanID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		adapter.httpClient,
	)
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			_ = httpResp.Body.Close()
		}
	}()
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
	plan, err := MapPlanEntityToCanonical(cfPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to map plan: %w", err)
	}
	return plan, nil
}
