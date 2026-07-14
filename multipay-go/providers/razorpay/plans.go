package razorpay

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// buildPlanCreateData builds the Razorpay Plan.Create request body from a canonical CreatePlanRequest.
func buildPlanCreateData(req *domain.CreatePlanRequest) (map[string]any, error) {
	body := &razorpayPlanCreateRequest{
		Period:   mapPlanIntervalTypeToRazorpay(req.IntervalType),
		Interval: req.Interval,
		Item: razorpayItem{
			Name:        req.PlanName,
			Amount:      int64(req.AmountMinor),
			Currency:    string(req.Currency),
			Description: req.Description,
		},
	}
	notes := map[string]string{}
	if req.Note != "" {
		notes["note"] = req.Note
	}
	if req.PlanID != "" {
		notes["plan_id"] = req.PlanID
	}
	if req.MaxCycles > 0 {
		notes["max_cycles"] = strconv.Itoa(int(req.MaxCycles))
	}
	if len(notes) > 0 {
		body.Notes = notes
	}
	return encodeRequest(body)
}

// createPlan creates a new subscription plan on the Razorpay payment gateway.
// Maps the canonical domain.CreatePlanRequest to a Razorpay SDK map format,
// calls the SDK, and maps the response back to a canonical domain.Plan.
func createPlan(ctx context.Context, a *Adapter, req *domain.CreatePlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}

	data, err := buildPlanCreateData(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build plan request: %w", err)
	}

	responseMap, err := a.client.Plan.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	typed, err := decodeResponse[razorpayPlanResponse](responseMap)
	if err != nil {
		return nil, err
	}

	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plan response: %w", err)
	}

	return mapPlanFromResponse(typed, rawJSON), nil
}

// getPlan retrieves an existing plan from the Razorpay payment gateway.
// Calls the Razorpay SDK to fetch the plan and maps it to a canonical domain.Plan.
func getPlan(ctx context.Context, a *Adapter, req *domain.GetPlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}

	if req.PlanID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch plan
	responseMap, err := a.client.Plan.Fetch(req.PlanID, nil, nil)
	if err != nil {
		// Check if plan not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("plan %s not found: %w", req.PlanID, domain.ErrProviderError)
		}
		return nil, fmt.Errorf("failed to fetch plan: %w", err)
	}

	typed, err := decodeResponse[razorpayPlanResponse](responseMap)
	if err != nil {
		return nil, err
	}

	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plan response: %w", err)
	}

	return mapPlanFromResponse(typed, rawJSON), nil
}
