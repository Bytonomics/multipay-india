package razorpay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// createSubscription creates a new subscription on Razorpay.
// If PlanDetails are provided, auto-creates the plan first, then creates the subscription with the auto-created plan_id.
// Otherwise, uses the provided PlanID.
// Maps the canonical domain.CreateSubscriptionRequest to Razorpay API calls,
// and maps the response back to a canonical domain.Subscription.
func createSubscription(ctx context.Context, adapter *Adapter, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	planID := req.PlanID

	// If PlanDetails are provided, auto-create the plan first
	if req.PlanDetails != nil {
		planData, err := buildPlanCreateData(req.PlanDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to build plan data: %w", err)
		}

		planResp, err := adapter.client.Plan.Create(planData, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create plan on razorpay: %w", domain.ErrProviderError)
		}

		if planResp == nil {
			return nil, fmt.Errorf("razorpay returned nil plan response: %w", domain.ErrProviderError)
		}

		// Decode plan response and extract plan ID
		planTyped, err := decodeResponse[razorpayPlanResponse](planResp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode plan response: %w", err)
		}
		planID = planTyped.ID
	}

	// Build subscription data
	subData := map[string]any{
		"plan_id":         planID,
		"customer_notify": 1,
	}

	// Add customer details
	if req.CustomerEmail != "" {
		subData["customer_email"] = req.CustomerEmail
	}
	if req.CustomerPhone != "" {
		subData["customer_phone"] = req.CustomerPhone
	}
	if req.CustomerName != "" {
		subData["customer_name"] = req.CustomerName
	}

	// Add optional fields
	if req.ExpiresAt != nil {
		subData["expire_by"] = req.ExpiresAt.Unix()
	}
	if req.FirstChargeTime != nil {
		subData["start_at"] = req.FirstChargeTime.Unix()
	}

	// Add tags as notes
	if len(req.Tags) > 0 {
		notesMap := make(map[string]any)
		for k, v := range req.Tags {
			notesMap[k] = v
		}
		subData["notes"] = notesMap
	}

	// Call Razorpay Subscription.Create()
	subResp, err := adapter.client.Subscription.Create(subData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription on razorpay: %w", domain.ErrProviderError)
	}

	if subResp == nil {
		return nil, fmt.Errorf("razorpay returned nil subscription response: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	typed, err := decodeResponse[razorpaySubscriptionResponse](subResp)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription response: %w", err)
	}
	return mapSubscriptionFromResponse(typed, rawJSON), nil
}

// getSubscription retrieves an existing subscription from Razorpay.
// Maps the canonical domain.GetSubscriptionRequest to a Razorpay fetch request,
// and maps the response back to a canonical domain.Subscription.
func getSubscription(ctx context.Context, adapter *Adapter, req *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Call Razorpay Subscription.Fetch()
	subResp, err := adapter.client.Subscription.Fetch(req.SubscriptionID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription from razorpay: %w", domain.ErrProviderError)
	}

	if subResp == nil {
		return nil, fmt.Errorf("razorpay returned nil subscription response: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	typed, err := decodeResponse[razorpaySubscriptionResponse](subResp)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription response: %w", err)
	}
	return mapSubscriptionFromResponse(typed, rawJSON), nil
}

// cancelSubscription cancels a subscription on Razorpay.
func cancelSubscription(ctx context.Context, adapter *Adapter, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Call Razorpay Subscription.Cancel()
	subResp, err := adapter.client.Subscription.Cancel(req.SubscriptionID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription on razorpay: %w", domain.ErrProviderError)
	}

	if subResp == nil {
		return nil, fmt.Errorf("razorpay returned nil subscription response: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	typed, err := decodeResponse[razorpaySubscriptionResponse](subResp)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription response: %w", err)
	}
	return mapSubscriptionFromResponse(typed, rawJSON), nil
}

// pauseSubscription pauses a subscription on Razorpay.
func pauseSubscription(ctx context.Context, adapter *Adapter, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Call Razorpay Subscription.Pause()
	subResp, err := adapter.client.Subscription.Pause(req.SubscriptionID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to pause subscription on razorpay: %w", domain.ErrProviderError)
	}

	if subResp == nil {
		return nil, fmt.Errorf("razorpay returned nil subscription response: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	typed, err := decodeResponse[razorpaySubscriptionResponse](subResp)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription response: %w", err)
	}
	return mapSubscriptionFromResponse(typed, rawJSON), nil
}

// resumeSubscription resumes a subscription on Razorpay.
func resumeSubscription(ctx context.Context, adapter *Adapter, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Call Razorpay Subscription.Resume()
	subResp, err := adapter.client.Subscription.Resume(req.SubscriptionID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to resume subscription on razorpay: %w", domain.ErrProviderError)
	}

	if subResp == nil {
		return nil, fmt.Errorf("razorpay returned nil subscription response: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	typed, err := decodeResponse[razorpaySubscriptionResponse](subResp)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription response: %w", err)
	}
	return mapSubscriptionFromResponse(typed, rawJSON), nil
}

// changePlan changes the plan of a subscription on Razorpay.
// Razorpay respects the ScheduleAt field:
// - ScheduleAt=NOW: applies change immediately (schedule_change_at: "now")
// - ScheduleAt=CYCLE_END: applies change at next billing cycle (schedule_change_at: "cycle_end")
func changePlan(ctx context.Context, adapter *Adapter, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Determine schedule_change_at based on ScheduleAt
	scheduleAt := "cycle_end" // default
	if req.ScheduleAt == domain.ScheduleChangeNow {
		scheduleAt = "now"
	}

	updateData := map[string]any{
		"plan_id":            req.NewPlanID,
		"schedule_change_at": scheduleAt,
	}

	// Call Razorpay Subscription.Update()
	subResp, err := adapter.client.Subscription.Update(req.SubscriptionID, updateData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to change plan on razorpay: %w", domain.ErrProviderError)
	}

	if subResp == nil {
		return nil, fmt.Errorf("razorpay returned nil subscription response: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	typed, err := decodeResponse[razorpaySubscriptionResponse](subResp)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription response: %w", err)
	}
	return mapSubscriptionFromResponse(typed, rawJSON), nil
}

// getSubscriptionPayments retrieves all payments for a subscription on Razorpay.
// Fetches invoices linked to the subscription and maps them to subscription payment records.
func getSubscriptionPayments(ctx context.Context, adapter *Adapter, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	resp, err := adapter.client.Invoice.All(map[string]any{"subscription_id": req.SubscriptionID}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription invoices: %w", err)
	}
	typed, err := decodeResponse[razorpayInvoiceListResponse](resp)
	if err != nil {
		return nil, err
	}
	payments := make([]*domain.SubscriptionPayment, 0, len(typed.Items))
	for i := range typed.Items {
		payments = append(payments, mapInvoiceToSubscriptionPayment(&typed.Items[i], req.SubscriptionID, nil))
	}
	return payments, nil
}

// Adapter method stubs - these delegate to the operation functions above

// CreateSubscription creates a new subscription.
// See subscriptions.go for implementation.
func (a *Adapter) CreateSubscription(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	return createSubscription(ctx, a, req)
}

// GetSubscription retrieves an existing subscription.
// See subscriptions.go for implementation.
func (a *Adapter) GetSubscription(ctx context.Context, req *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	return getSubscription(ctx, a, req)
}

// CancelSubscription cancels a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) CancelSubscription(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	return cancelSubscription(ctx, a, req)
}

// PauseSubscription pauses a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) PauseSubscription(ctx context.Context, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	return pauseSubscription(ctx, a, req)
}

// ResumeSubscription resumes a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) ResumeSubscription(ctx context.Context, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	return resumeSubscription(ctx, a, req)
}

// ChangePlan changes the plan of a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) ChangePlan(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	return changePlan(ctx, a, req)
}

// GetSubscriptionPayments retrieves all payments for a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) GetSubscriptionPayments(ctx context.Context, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	return getSubscriptionPayments(ctx, a, req)
}
