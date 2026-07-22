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

	// Build subscription data as typed struct
	subData := &razorpaySubscriptionCreateRequest{
		PlanID:        planID,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		CustomerName:  req.CustomerName,
	}
	// customer_notify: forward the caller's choice verbatim (true→1, false→0). The library imposes
	// NO default — nil leaves it omitted so Razorpay applies its own default.
	if req.CustomerNotify != nil {
		notify := 0
		if *req.CustomerNotify {
			notify = 1
		}
		subData.CustomerNotify = &notify
	}

	// total_count is MANDATORY for Razorpay (unless end_at is used) — send it unconditionally.
	// Prefer the top-level canonical TotalCount (which the existing-PlanID path can supply);
	// fall back to the inline plan's MaxCycles. When neither bounds the cycles, it is omitted
	// (omitempty) — the caller must then supply total_count via TotalCount or Razorpay rejects
	// the subscription.
	switch {
	case req.TotalCount > 0:
		subData.TotalCount = int64(req.TotalCount)
	case req.PlanDetails != nil && req.PlanDetails.MaxCycles > 0:
		subData.TotalCount = int64(req.PlanDetails.MaxCycles)
	}

	// Add optional fields
	if req.Quantity > 0 {
		subData.Quantity = req.Quantity
	}
	if req.OfferID != "" {
		subData.OfferID = req.OfferID
	}
	if len(req.Addons) > 0 {
		addons := make([]razorpayAddon, 0, len(req.Addons))
		for i := range req.Addons {
			addons = append(addons, razorpayAddon{
				Item: razorpayItem{
					Name:     req.Addons[i].Name,
					Amount:   int64(req.Addons[i].AmountMinor), // Razorpay native minor units
					Currency: string(req.Addons[i].Currency),
				},
			})
		}
		subData.Addons = addons
	}
	if req.ExpiresAt != nil {
		subData.ExpireBy = req.ExpiresAt.Unix()
	}
	if req.FirstChargeTime != nil {
		subData.StartAt = req.FirstChargeTime.Unix()
	}

	// Add tags as notes
	if len(req.Tags) > 0 {
		notes := make(map[string]any)
		for k, v := range req.Tags {
			notes[k] = v
		}
		subData.Notes = notes
	}

	// Convert typed struct to map for SDK
	subDataMap, err := encodeRequest(subData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode subscription request: %w", err)
	}

	// Call Razorpay Subscription.Create()
	subResp, err := adapter.client.Subscription.Create(subDataMap, nil)
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
	return mapSubscriptionFromResponse(adapter.config.Environment, typed, rawJSON), nil
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
	return mapSubscriptionFromResponse(adapter.config.Environment, typed, rawJSON), nil
}

// cancelSubscription cancels a subscription on Razorpay.
func cancelSubscription(ctx context.Context, adapter *Adapter, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Forward the optional cancel_at_cycle_end body param (Razorpay Cancel Subscription).
	// nil → immediate cancel (nil data map); non-nil → cancel per the requested timing.
	var cancelData map[string]any
	if req.CancelAtCycleEnd != nil {
		cancelData = map[string]any{"cancel_at_cycle_end": *req.CancelAtCycleEnd}
	}

	// Call Razorpay Subscription.Cancel()
	subResp, err := adapter.client.Subscription.Cancel(req.SubscriptionID, cancelData, nil)
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
	return mapSubscriptionFromResponse(adapter.config.Environment, typed, rawJSON), nil
}

// pauseSubscription pauses a subscription on Razorpay.
func pauseSubscription(ctx context.Context, adapter *Adapter, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Forward the optional pause_at body param (only "now" is accepted; validated at the boundary).
	var pauseData map[string]any
	if req.PauseAt != "" {
		pauseData = map[string]any{"pause_at": req.PauseAt}
	}

	// Call Razorpay Subscription.Pause()
	subResp, err := adapter.client.Subscription.Pause(req.SubscriptionID, pauseData, nil)
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
	return mapSubscriptionFromResponse(adapter.config.Environment, typed, rawJSON), nil
}

// resumeSubscription resumes a subscription on Razorpay.
func resumeSubscription(ctx context.Context, adapter *Adapter, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Forward the optional resume_at body param (only "now" is accepted; validated at the boundary).
	var resumeData map[string]any
	if req.ResumeAt != "" {
		resumeData = map[string]any{"resume_at": req.ResumeAt}
	}

	// Call Razorpay Subscription.Resume()
	subResp, err := adapter.client.Subscription.Resume(req.SubscriptionID, resumeData, nil)
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
	return mapSubscriptionFromResponse(adapter.config.Environment, typed, rawJSON), nil
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

	// Build change plan request as typed struct
	updateReq := &razorpayChangePlanRequest{
		PlanID:           req.NewPlanID,
		ScheduleChangeAt: scheduleAt,
		OfferID:          req.OfferID,
		Quantity:         req.Quantity,
		RemainingCount:   req.RemainingCount,
		CustomerNotify:   req.CustomerNotify,
	}
	if req.StartAt != nil {
		updateReq.StartAt = req.StartAt.Unix()
	}

	// Convert typed struct to map for SDK
	updateData, err := encodeRequest(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode change plan request: %w", err)
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
	return mapSubscriptionFromResponse(adapter.config.Environment, typed, rawJSON), nil
}

// getSubscriptionPayments retrieves all payments for a subscription on Razorpay.
// Fetches invoices linked to the subscription and maps them to subscription payment records.
func getSubscriptionPayments(ctx context.Context, adapter *Adapter, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Build invoice list request as typed struct
	invoiceReq := &razorpayInvoiceListRequest{
		SubscriptionID: req.SubscriptionID,
	}

	// Convert typed struct to map for SDK
	invoiceReqMap, err := encodeRequest(invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode invoice list request: %w", err)
	}

	resp, err := adapter.client.Invoice.All(invoiceReqMap, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription invoices: %w", err)
	}
	typed, err := decodeResponse[razorpayInvoiceListResponse](resp)
	if err != nil {
		return nil, err
	}
	payments := make([]*domain.SubscriptionPayment, 0, len(typed.Items))
	for i := range typed.Items {
		rawJSON, err := json.Marshal(&typed.Items[i])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal invoice: %w", err)
		}
		payments = append(payments, mapInvoiceToSubscriptionPayment(&typed.Items[i], req.SubscriptionID, rawJSON))
	}
	return payments, nil
}

// chargeSubscription performs an on-demand charge on a subscription using the CreateAddon method.
// Maps the canonical domain.ChargeSubscriptionRequest to Razorpay addon data,
// calls the SDK, and maps the response back to a canonical domain.SubscriptionPayment.
func chargeSubscription(ctx context.Context, adapter *Adapter, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Build addon request data as typed struct.
	// NOTE: item.name is a mandatory Razorpay addon field. We source it from Remarks; if the
	// caller leaves Remarks empty, Razorpay will reject the addon (mandatory-field risk). The
	// canonical ChargeSubscriptionRequest keeps Remarks optional because Cashfree's charge path
	// does not require it — callers targeting Razorpay MUST supply Remarks.
	addonData := &razorpayAddonRequest{
		Item: razorpayItem{
			Name:        req.Remarks,
			Amount:      int64(req.AmountMinor), // Razorpay native minor units
			Currency:    string(req.Currency),
			Description: req.Description,
		},
		Quantity: 1,
	}

	// Convert typed struct to map for SDK
	addonDataMap, err := encodeRequest(addonData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode addon request: %w", err)
	}

	// Call Razorpay Subscription.CreateAddon()
	addonResp, err := adapter.client.Subscription.CreateAddon(req.SubscriptionID, addonDataMap, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create addon on razorpay: %w", domain.ErrProviderError)
	}

	if addonResp == nil {
		return nil, fmt.Errorf("razorpay returned nil addon response: %w", domain.ErrProviderError)
	}

	// Decode response to typed struct - try as invoice-like payment response
	typed, err := decodeResponse[razorpayInvoiceResponse](addonResp)
	if err != nil {
		return nil, err
	}

	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal addon response: %w", err)
	}

	// Map invoice response to subscription payment
	payment := mapInvoiceToSubscriptionPayment(typed, req.SubscriptionID, rawJSON)
	if payment == nil {
		return nil, fmt.Errorf("failed to map addon response to subscription payment: %w", domain.ErrProviderError)
	}

	return payment, nil
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

// ChargeSubscription performs an on-demand charge on a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) ChargeSubscription(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	return chargeSubscription(ctx, a, req)
}
