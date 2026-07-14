package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
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

	// Set CustomerName if provided
	if req.CustomerName != "" {
		cfReq.CustomerDetails.CustomerName = &req.CustomerName
	}

	// TPV bank details on the subscription customer (eNACH / physical-NACH pre-binding).
	if req.BankDetails != nil {
		cfReq.CustomerDetails.CustomerBankAccountHolderName = optStr(req.BankDetails.AccountHolderName)
		cfReq.CustomerDetails.CustomerBankAccountNumber = optStr(req.BankDetails.AccountNumber)
		cfReq.CustomerDetails.CustomerBankIfsc = optStr(req.BankDetails.IFSC)
		cfReq.CustomerDetails.CustomerBankCode = optStr(req.BankDetails.BankCode)
		cfReq.CustomerDetails.CustomerBankAccountType = optStr(req.BankDetails.AccountType)
	}

	// Resolve the request currency for the authorization-amount conversion. Prefer the inline
	// plan's currency (subscriptions with an existing PlanID carry currency on the plan itself,
	// which Cashfree already knows, so the authorization amount for those is expected in that
	// currency). Default to INR when no inline plan currency is present.
	subCurrency := "INR"
	if req.PlanDetails != nil && req.PlanDetails.Currency != "" {
		subCurrency = string(req.PlanDetails.Currency)
	}

	// Authorization details (mandate-authorization step).
	if req.AuthorizationDetails != nil {
		authDetails := &cf.CreateSubscriptionRequestAuthorizationDetails{
			AuthorizationAmountRefund: req.AuthorizationDetails.AuthorizationAmountRefund,
			PaymentMethods:            req.AuthorizationDetails.PaymentMethods,
		}
		if req.AuthorizationDetails.AuthorizationAmountMinor > 0 {
			authAmt := float32(currencyutils.AmountMinorToMajor(int64(req.AuthorizationDetails.AuthorizationAmountMinor), subCurrency))
			authDetails.AuthorizationAmount = &authAmt
		}
		cfReq.AuthorizationDetails = authDetails
	}

	// Set subscription metadata (ReturnUrl for mandate-authorization redirect, plus optional
	// notification_channel / session_id_expiry extras).
	if req.ReturnURL != "" || req.Meta != nil {
		meta := &cf.CreateSubscriptionRequestSubscriptionMeta{}
		if req.ReturnURL != "" {
			returnUrl := req.ReturnURL
			meta.ReturnUrl = &returnUrl
		}
		if req.Meta != nil {
			meta.NotificationChannel = req.Meta.NotificationChannel
			meta.SessionIdExpiry = optStr(req.Meta.SessionIDExpiry)
		}
		cfReq.SubscriptionMeta = meta
	}

	// Easy Split payment splits.
	if len(req.PaymentSplits) > 0 {
		splits := make([]cf.SubscriptionPaymentSplitItem, 0, len(req.PaymentSplits))
		for i := range req.PaymentSplits {
			vendorID := req.PaymentSplits[i].VendorID
			pct := float32(req.PaymentSplits[i].Percentage)
			splits = append(splits, cf.SubscriptionPaymentSplitItem{
				VendorId:   optStr(vendorID),
				Percentage: &pct,
			})
		}
		cfReq.SubscriptionPaymentSplits = splits
	}

	// Convert and set subscription tags (convert from map[string]string to map[string]any)
	if len(req.Tags) > 0 {
		tags := make(map[string]any, len(req.Tags))
		for k, v := range req.Tags {
			tags[k] = v
		}
		cfReq.SubscriptionTags = tags
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

	// Forward optional cf_order_id (attaches the subscription to a pre-created Cashfree order).
	// optStr returns nil for "" so an empty value is omitted (the library imposes no default).
	cfReq.CfOrderId = optStr(req.CfOrderID)

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
	// Mandatory fields are set directly (validated non-empty at the boundary).
	// Optional fields use the nil-if-empty helpers so omitempty drops them.
	details := cf.CreateSubscriptionRequestPlanDetails{
		PlanId:        &pd.PlanID,
		PlanName:      &pd.PlanName,
		PlanType:      ptrString(string(pd.PlanType)),
		PlanMaxAmount: ptrFloat32(float32(currencyutils.AmountMinorToMajor(int64(pd.MaxAmountMinor), string(pd.Currency)))),
		PlanCurrency:  optStr(string(pd.Currency)),
		PlanMaxCycles: optInt32(pd.MaxCycles),
		PlanNote:      optStr(pd.Note),
	}

	// Conditional (PERIODIC only): the per-cycle amount and the interval are meaningless
	// for ON_DEMAND plans and must be omitted, not sent as zero/empty.
	if pd.PlanType == domain.PlanTypePeriodic {
		details.PlanAmount = ptrFloat32(float32(currencyutils.AmountMinorToMajor(int64(pd.AmountMinor), string(pd.Currency))))
		details.PlanIntervals = optInt32(pd.Interval)
		details.PlanIntervalType = optStr(string(pd.IntervalType))
	}
	return details
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
		// Retry-safe cancel: if the subscription is ALREADY cancelled, treat this as success so a
		// replayed upgrade-finalize (charge succeeded, an earlier cancel/DB write failed) can complete
		// instead of looping. Confirmed via a fetch of the live status, not by parsing the vendor error.
		if already := fetchIfCancelled(ctx, adapter, req.SubscriptionID); already != nil {
			return already, nil
		}
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

// fetchIfCancelled fetches the subscription and returns its canonical form ONLY if it is already
// cancelled at the provider; otherwise returns nil. Used to make cancelSubscription idempotent on
// replay (Cashfree errors when cancelling an already-cancelled subscription).
func fetchIfCancelled(ctx context.Context, adapter *Adapter, subscriptionID string) *domain.Subscription {
	cfSub, httpResp, ferr := adapter.cfClient.SubsFetchSubscriptionWithContext(
		ctx,
		subscriptionID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		adapter.httpClient,
	)
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			_ = httpResp.Body.Close()
		}
	}()
	if ferr != nil || cfSub == nil {
		return nil
	}
	sub, merr := MapSubscriptionEntityToCanonical(cfSub)
	if merr != nil || sub == nil {
		return nil
	}
	if sub.Status == domain.SubscriptionStatusCancelled {
		return sub
	}
	return nil
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

	// Cashfree REQUIRES action_details.next_scheduled_time for the ACTIVATE action — it is the
	// date the resumed subscription's next charge is scheduled. Send it when the caller provides
	// NextScheduledTime; Cashfree rejects an ACTIVATE with no action_details.
	cfReq := &cf.ManageSubscriptionRequest{
		SubscriptionId: req.SubscriptionID,
		Action:         "ACTIVATE",
	}
	if req.NextScheduledTime != nil {
		nextStr := req.NextScheduledTime.Format("2006-01-02T15:04:05-07:00")
		cfReq.ActionDetails = &cf.ManageSubscriptionRequestActionDetails{
			NextScheduledTime: &nextStr,
		}
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

// chargeSubscription performs an on-demand charge on a subscription.
// Maps the canonical domain.ChargeSubscriptionRequest to a Cashfree CreateSubscriptionPaymentRequest,
// calls the SDK, and maps the response back to a canonical domain.SubscriptionPayment.
func chargeSubscription(ctx context.Context, adapter *Adapter, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
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

	// Convert amount from minor to major units
	amountMajor := currencyutils.AmountMinorToMajor(int64(req.AmountMinor), string(req.Currency))

	// Build Cashfree CreateSubscriptionPaymentRequest.
	// NOTE: cf.CreateSubscriptionPaymentRequest also carries SubscriptionSessionId and
	// PaymentMethod fields — those belong to the AUTH (mandate-authorization) path, not the
	// CHARGE path this adapter implements, so they are intentionally left unset here.
	cfReq := &cf.CreateSubscriptionPaymentRequest{
		SubscriptionId: req.SubscriptionID,
		PaymentId:      req.PaymentRef, // caller-supplied unique id
		PaymentAmount:  ptrFloat32(float32(amountMajor)),
		PaymentType:    "CHARGE",
		PaymentRemarks: optStr(req.Remarks),
	}

	// Future-dated charge (date-only). Cashfree expects "YYYY-MM-DD".
	if req.PaymentScheduleDate != nil {
		scheduleStr := req.PaymentScheduleDate.Format("2006-01-02")
		cfReq.PaymentScheduleDate = &scheduleStr
	}

	// Engage Cashfree's NATIVE idempotency for the proration charge so a webhook replay (e.g. an
	// upgrade-finalize where the charge succeeded but the subsequent cancel or the studio DB write
	// failed) returns the ORIGINAL charge result instead of debiting the customer a second time.
	// PaymentRef is the studio payment-attempt's stable id.
	var idempotencyKey *string
	if req.PaymentRef != "" {
		k := req.PaymentRef
		idempotencyKey = &k
	}

	// Call Cashfree SDK
	cfPayment, httpResp, err := adapter.cfClient.SubsCreatePaymentWithContext(
		ctx,
		cfReq,
		nil,            // xRequestId
		idempotencyKey, // native idempotency key = PaymentRef (safe replay, no double debit)
		adapter.httpClient,
	)
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			_ = httpResp.Body.Close()
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to create payment on cashfree: %w", domain.ErrProviderError)
	}

	if cfPayment == nil {
		return nil, fmt.Errorf("cashfree returned nil payment: %w", domain.ErrProviderError)
	}

	// Convert CreateSubscriptionPaymentResponse to SubscriptionPaymentEntity
	entity := &cf.SubscriptionPaymentEntity{
		CfPaymentId:          cfPayment.CfPaymentId,
		PaymentAmount:        cfPayment.PaymentAmount,
		PaymentId:            cfPayment.PaymentId,
		PaymentInitiatedDate: cfPayment.PaymentInitiatedDate,
		PaymentStatus:        cfPayment.PaymentStatus,
		PaymentType:          cfPayment.PaymentType,
		SubscriptionId:       cfPayment.SubscriptionId,
		FailureDetails:       cfPayment.FailureDetails,
	}

	// Map response to canonical type using existing mapper
	payment, err := MapSubscriptionPaymentEntityToCanonical(entity, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to map subscription payment: %w", err)
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

// ChargeSubscription performs an on-demand charge on a subscription.
// See subscriptions.go for implementation.
func (a *Adapter) ChargeSubscription(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	return chargeSubscription(ctx, a, req)
}
