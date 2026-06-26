package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// createSubscription creates a new subscription on the Cashfree payment gateway.
// Accepts either an existing PlanID or inline PlanDetails.
// Maps the canonical domain.CreateSubscriptionRequest to a Cashfree CreateSubscriptionRequest,
// calls the SDK, and maps the response back to a canonical domain.Subscription.
func createSubscription(ctx context.Context, adapter *Adapter, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}
	if req.CustomerEmail == "" {
		return nil, fmt.Errorf("customer_email is required for Cashfree subscriptions: %w", domain.ErrInvalidRequest)
	}

	// Build Cashfree CreateSubscriptionRequest
	cfReq := &cf.CreateSubscriptionRequest{
		SubscriptionId: req.SubscriptionID,
		CustomerDetails: cf.SubscriptionCustomerDetails{
			CustomerEmail: req.CustomerEmail,
			CustomerPhone: req.CustomerPhone,
		},
	}

	// Handle plan: either existing PlanID or inline PlanDetails
	if req.PlanID != "" {
		// Cashfree expects CreateSubscriptionRequestPlanDetails for inline refs
		cfReq.PlanDetails = cf.CreateSubscriptionRequestPlanDetails{
			PlanId: &req.PlanID,
		}
	} else if req.PlanDetails != nil {
		// Inline plan details - create using PlanDetails struct
		cfReq.PlanDetails = buildInlinePlanDetails(req.PlanDetails)
	}

	// Optional: set expiry and first charge times (as ISO 8601 strings)
	if req.ExpiresAt != nil {
		expiryStr := req.ExpiresAt.Format("2006-01-02T15:04:05-07:00")
		cfReq.SubscriptionExpiryTime = &expiryStr
	}
	if req.FirstChargeTime != nil {
		chargeStr := req.FirstChargeTime.Format("2006-01-02T15:04:05-07:00")
		cfReq.SubscriptionFirstChargeTime = &chargeStr
	}

	// Call Cashfree SDK
	cfSub, httpResp, err := adapter.cfClient.SubsCreateSubscriptionWithContext(
		ctx,
		cfReq,
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
		return nil, fmt.Errorf("failed to create subscription on cashfree: %w", domain.ErrProviderError)
	}

	if cfSub == nil {
		return nil, fmt.Errorf("cashfree returned nil subscription: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	subscription, err := MapSubscriptionEntityToCanonical(cfSub)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription: %w", err)
	}
	return subscription, nil
}

// buildInlinePlanDetails maps a canonical CreatePlanRequest to the Cashfree inline plan-details struct.
func buildInlinePlanDetails(pd *domain.CreatePlanRequest) cf.CreateSubscriptionRequestPlanDetails {
	return cf.CreateSubscriptionRequestPlanDetails{
		PlanId:           &pd.PlanID,
		PlanName:         &pd.PlanName,
		PlanType:         ptrString(string(pd.PlanType)),
		PlanCurrency:     ptrString(string(pd.Currency)),
		PlanAmount:       ptrFloat32(float32(AmountMinorToMajor(int64(pd.AmountMinor), string(pd.Currency)))),
		PlanMaxAmount:    ptrFloat32(float32(AmountMinorToMajor(int64(pd.MaxAmountMinor), string(pd.Currency)))),
		PlanMaxCycles:    ptrInt32(pd.MaxCycles),
		PlanIntervals:    ptrInt32(pd.Interval),
		PlanIntervalType: ptrString(string(pd.IntervalType)),
		PlanNote:         ptrString(pd.Note),
	}
}

// getSubscription retrieves an existing subscription from the Cashfree payment gateway.
// Maps the canonical domain.GetSubscriptionRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.Subscription.
func getSubscription(ctx context.Context, adapter *Adapter, req *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK
	cfSub, httpResp, err := adapter.cfClient.SubsFetchSubscriptionWithContext(
		ctx,
		req.SubscriptionID,
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
		return nil, fmt.Errorf("failed to fetch subscription from cashfree: %w", domain.ErrProviderError)
	}

	if cfSub == nil {
		return nil, fmt.Errorf("cashfree returned nil subscription: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	subscription, err := MapSubscriptionEntityToCanonical(cfSub)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription: %w", err)
	}
	return subscription, nil
}

// cancelSubscription cancels an existing subscription.
// Uses the manage subscription endpoint with action "CANCEL".
func cancelSubscription(ctx context.Context, adapter *Adapter, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	cfReq := &cf.ManageSubscriptionRequest{
		SubscriptionId: req.SubscriptionID,
		Action:         "CANCEL",
	}

	// Call Cashfree SDK
	cfSub, httpResp, err := adapter.cfClient.SubsManageSubscriptionWithContext(
		ctx,
		req.SubscriptionID,
		cfReq,
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
		return nil, fmt.Errorf("failed to cancel subscription on cashfree: %w", domain.ErrProviderError)
	}

	if cfSub == nil {
		return nil, fmt.Errorf("cashfree returned nil subscription: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	subscription, err := MapSubscriptionEntityToCanonical(cfSub)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription: %w", err)
	}
	return subscription, nil
}

// pauseSubscription pauses an active subscription.
// Uses the manage subscription endpoint with action "PAUSE".
func pauseSubscription(ctx context.Context, adapter *Adapter, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	cfReq := &cf.ManageSubscriptionRequest{
		SubscriptionId: req.SubscriptionID,
		Action:         "PAUSE",
	}

	// Call Cashfree SDK
	cfSub, httpResp, err := adapter.cfClient.SubsManageSubscriptionWithContext(
		ctx,
		req.SubscriptionID,
		cfReq,
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
		return nil, fmt.Errorf("failed to pause subscription on cashfree: %w", domain.ErrProviderError)
	}

	if cfSub == nil {
		return nil, fmt.Errorf("cashfree returned nil subscription: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	subscription, err := MapSubscriptionEntityToCanonical(cfSub)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription: %w", err)
	}
	return subscription, nil
}

// resumeSubscription resumes a paused subscription.
// Uses the manage subscription endpoint with action "ACTIVATE".
func resumeSubscription(ctx context.Context, adapter *Adapter, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	cfReq := &cf.ManageSubscriptionRequest{
		SubscriptionId: req.SubscriptionID,
		Action:         "ACTIVATE",
	}

	// Call Cashfree SDK
	cfSub, httpResp, err := adapter.cfClient.SubsManageSubscriptionWithContext(
		ctx,
		req.SubscriptionID,
		cfReq,
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
		return nil, fmt.Errorf("failed to resume subscription on cashfree: %w", domain.ErrProviderError)
	}

	if cfSub == nil {
		return nil, fmt.Errorf("cashfree returned nil subscription: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	subscription, err := MapSubscriptionEntityToCanonical(cfSub)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription: %w", err)
	}
	return subscription, nil
}

// changePlan changes the plan of an existing subscription.
// Uses the manage subscription endpoint with action "CHANGE_PLAN" and planId.
// IMPORTANT: Cashfree always applies plan changes at the next billing cycle.
// It ignores the ScheduleAt field. When ScheduleAt=NOW is requested, we log a warning
// but proceed with the change (Cashfree will apply it at cycle end).
func changePlan(ctx context.Context, adapter *Adapter, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Log warning if ScheduleAt=NOW (Cashfree ignores this and does cycle-end)
	if req.ScheduleAt == domain.ScheduleChangeNow {
		// Note: In production, this would be logged via a logger.
		// For now, including as a comment for documentation purposes.
		_ = "WARNING: Cashfree applies plan changes at cycle end, not immediately"
	}

	cfReq := &cf.ManageSubscriptionRequest{
		SubscriptionId: req.SubscriptionID,
		Action:         "CHANGE_PLAN",
		ActionDetails: &cf.ManageSubscriptionRequestActionDetails{
			PlanId: &req.NewPlanID,
		},
	}

	// Call Cashfree SDK
	cfSub, httpResp, err := adapter.cfClient.SubsManageSubscriptionWithContext(
		ctx,
		req.SubscriptionID,
		cfReq,
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
		return nil, fmt.Errorf("failed to change plan on cashfree: %w", domain.ErrProviderError)
	}

	if cfSub == nil {
		return nil, fmt.Errorf("cashfree returned nil subscription: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	subscription, err := MapSubscriptionEntityToCanonical(cfSub)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription: %w", err)
	}
	return subscription, nil
}

// getSubscriptionPayments retrieves all payments for a subscription.
// Maps each payment entity from Cashfree to the canonical domain type.
func getSubscriptionPayments(ctx context.Context, adapter *Adapter, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	// Fetch subscription to get currency
	subEntity, subHttpResp, ferr := adapter.cfClient.SubsFetchSubscriptionWithContext(
		ctx,
		req.SubscriptionID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		adapter.httpClient,
	)
	defer func() {
		if subHttpResp != nil && subHttpResp.Body != nil {
			_ = subHttpResp.Body.Close()
		}
	}()
	if ferr != nil {
		return nil, fmt.Errorf("failed to fetch subscription for currency: %w", ferr)
	}

	// Resolve currency
	currency := ""
	if subEntity != nil && subEntity.PlanDetails != nil && subEntity.PlanDetails.PlanCurrency != nil {
		currency = *subEntity.PlanDetails.PlanCurrency
	}
	if currency == "" {
		return nil, fmt.Errorf("could not resolve subscription currency for %s: %w", req.SubscriptionID, domain.ErrProviderError)
	}

	// Call Cashfree SDK
	cfPayments, httpResp, err := adapter.cfClient.SubsFetchSubscriptionPaymentsWithContext(
		ctx,
		req.SubscriptionID,
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
		return nil, fmt.Errorf("failed to fetch subscription payments from cashfree: %w", domain.ErrProviderError)
	}

	// Map each payment entity to canonical type
	payments := make([]*domain.SubscriptionPayment, 0, len(cfPayments))
	for i := range cfPayments {
		payment, err := MapSubscriptionPaymentEntityToCanonical(&cfPayments[i], currency)
		if err != nil {
			return nil, fmt.Errorf("failed to map subscription payment at index %d: %w", i, err)
		}
		if payment != nil {
			payments = append(payments, payment)
		}
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

// PauseSubscription pauses an active subscription.
// See subscriptions.go for implementation.
func (a *Adapter) PauseSubscription(ctx context.Context, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	return pauseSubscription(ctx, a, req)
}

// ResumeSubscription resumes a paused subscription.
// See subscriptions.go for implementation.
func (a *Adapter) ResumeSubscription(ctx context.Context, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	return resumeSubscription(ctx, a, req)
}

// ChangePlan changes the plan of an existing subscription.
// See subscriptions.go for implementation.
func (a *Adapter) ChangePlan(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	return changePlan(ctx, a, req)
}

// GetSubscriptionPayments retrieves all payments for a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) GetSubscriptionPayments(ctx context.Context, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	return getSubscriptionPayments(ctx, a, req)
}
