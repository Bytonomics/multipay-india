package cashfree

import (
	"context"
	"fmt"
	"time"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

// createPaymentLink creates a new shareable payment link on the Cashfree payment gateway.
// Maps the canonical domain.CreatePaymentLinkRequest to a Cashfree CreateLinkRequest,
// calls the SDK, and maps the response back to a canonical domain.PaymentLink.
func createPaymentLink(ctx context.Context, adapter *Adapter, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.AmountMinor <= 0 {
		return nil, fmt.Errorf("amount must be positive: %w", domain.ErrInvalidRequest)
	}

	if req.Currency == "" {
		return nil, fmt.Errorf("currency is required: %w", domain.ErrInvalidRequest)
	}

	// Build Cashfree CreateLinkRequest
	purpose := req.Purpose

	customerDetails := cf.LinkCustomerDetailsEntity{
		CustomerPhone: req.Customer.Phone,
	}
	// Guard optional customer email/name pointers
	if req.Customer.Email != "" {
		email := req.Customer.Email
		customerDetails.CustomerEmail = &email
	}
	if req.Customer.Name != "" {
		name := req.Customer.Name
		customerDetails.CustomerName = &name
	}
	// Forward optional TPV bank fields on the link customer.
	if req.Customer.BankAccountNumber != "" {
		v := req.Customer.BankAccountNumber
		customerDetails.CustomerBankAccountNumber = &v
	}
	if req.Customer.BankIFSC != "" {
		v := req.Customer.BankIFSC
		customerDetails.CustomerBankIfsc = &v
	}
	if req.Customer.BankCode != 0 {
		v := req.Customer.BankCode
		customerDetails.CustomerBankCode = &v
	}
	// Forward optional TPV bank-account-holder name (cf field is misspelled on the wire).
	if req.Customer.BankAccountHolderName != "" {
		v := req.Customer.BankAccountHolderName
		customerDetails.CustomerBankAcoountHolderName = &v
	}

	cfReq := &cf.CreateLinkRequest{
		LinkAmount:          currencyutils.AmountMinorToMajor(int64(req.AmountMinor), string(req.Currency)),
		LinkCurrency:        string(req.Currency),
		LinkPurpose:         &purpose,
		CustomerDetails:     customerDetails,
		LinkPartialPayments: req.PartialPayment,
		LinkExpiryTime:      toRFC3339String(req.ExpiryTime),
		LinkAutoReminders:   req.AutoReminders,
	}

	// Forward optional link_minimum_partial_amount (minor → major).
	if req.MinPartialAmount != 0 {
		v := currencyutils.AmountMinorToMajor(int64(req.MinPartialAmount), string(req.Currency))
		cfReq.LinkMinimumPartialAmount = &v
	}

	// Forward optional order_splits (Easy Split) on the link.
	if len(req.OrderSplits) > 0 {
		cfReq.OrderSplits = mapVendorSplits(req.OrderSplits, string(req.Currency))
	}

	// Build LinkMeta from return_url / notify_url / upi_intent / payment_methods.
	// A single LinkMeta carries all four; only allocate it if at least one is set.
	if req.ReturnURL != "" || req.NotifyURL != "" || req.UpiIntent != "" || req.PaymentMethods != "" {
		linkMeta := &cf.LinkMetaResponseEntity{}
		if req.ReturnURL != "" {
			v := req.ReturnURL
			linkMeta.ReturnUrl = &v
		}
		if req.NotifyURL != "" {
			v := req.NotifyURL
			linkMeta.NotifyUrl = &v
		}
		if req.UpiIntent != "" {
			v := req.UpiIntent
			linkMeta.UpiIntent = &v
		}
		if req.PaymentMethods != "" {
			v := req.PaymentMethods
			linkMeta.PaymentMethods = &v
		}
		cfReq.LinkMeta = linkMeta
	}

	// Guard optional LinkId
	if req.LinkID != "" {
		linkId := req.LinkID
		cfReq.LinkId = &linkId
	}

	// Guard optional LinkNotes (only if Metadata non-empty)
	if len(req.Metadata) > 0 {
		metadataPtr := (map[string]string)(req.Metadata)
		cfReq.LinkNotes = &metadataPtr
	}

	// Guard LinkNotify (only if either notify flag is set)
	if req.NotifySMS != nil || req.NotifyEmail != nil {
		cfReq.LinkNotify = &cf.LinkNotifyEntity{
			SendSms:   req.NotifySMS,
			SendEmail: req.NotifyEmail,
		}
	}

	// Forward optional enable_invoice (Cashfree invoice generation for the link). Forwarded exactly
	// as the caller set it; nil ⇒ not sent (the library imposes no default).
	if req.EnableInvoice != nil {
		cfReq.EnableInvoice = req.EnableInvoice
	}

	// Forward optional link subscription (attaches a subscription mandate to the link).
	if req.Subscription != nil {
		cfReq.Subscription = mapLinkSubscription(req.Subscription, string(req.Currency))
	}

	// Call Cashfree SDK
	cfLink, httpResp, err := adapter.cfClient.PGCreateLinkWithContext(
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
		return nil, fmt.Errorf("failed to create payment link on cashfree: %w", domain.ErrProviderError)
	}

	if cfLink == nil {
		return nil, fmt.Errorf("cashfree returned nil payment link: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	link, err := MapLinkEntityToCanonical(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to map link: %w", err)
	}
	return link, nil
}

// getPaymentLink retrieves an existing payment link from the Cashfree payment gateway.
// Maps the canonical domain.GetPaymentLinkRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.PaymentLink.
func getPaymentLink(ctx context.Context, adapter *Adapter, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.LinkID == "" {
		return nil, fmt.Errorf("LinkID is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK to fetch payment link
	cfLink, httpResp, err := adapter.cfClient.PGFetchLinkWithContext(
		ctx,
		req.LinkID,
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
		// Check if error is 404 link not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment link %s not found: %w", req.LinkID, domain.ErrPaymentLinkNotFound)
		}
		return nil, fmt.Errorf("failed to fetch payment link from Cashfree: %w", domain.ErrProviderError)
	}

	if cfLink == nil {
		return nil, fmt.Errorf("cashfree returned nil payment link: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	link, err := MapLinkEntityToCanonical(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to map link: %w", err)
	}
	return link, nil
}

// cancelPaymentLink cancels an existing payment link on the Cashfree payment gateway.
// Maps the canonical domain.CancelPaymentLinkRequest to a Cashfree cancel request,
// calls the SDK, and maps the response back to a canonical domain.PaymentLink.
func cancelPaymentLink(ctx context.Context, adapter *Adapter, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.LinkID == "" {
		return nil, fmt.Errorf("LinkID is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK to cancel payment link
	cfLink, httpResp, err := adapter.cfClient.PGCancelLinkWithContext(
		ctx,
		req.LinkID,
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
		return nil, fmt.Errorf("failed to cancel payment link on cashfree: %w", domain.ErrProviderError)
	}

	if cfLink == nil {
		return nil, fmt.Errorf("cashfree returned nil payment link: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	link, err := MapLinkEntityToCanonical(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to map link: %w", err)
	}
	return link, nil
}

// toRFC3339String converts a time.Time pointer to an RFC3339-formatted string pointer.
// Returns nil if the input time is nil or zero-valued.
func toRFC3339String(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	rfc3339str := t.Format(time.RFC3339)
	return &rfc3339str
}

// mapLinkSubscription maps the canonical LinkSubscription to Cashfree's link `subscription` object.
// The authorization amount is converted minor→major via the LINK currency; plan_details amounts are
// converted via the PLAN's own currency (inside mapLinkPlanDetails).
func mapLinkSubscription(s *domain.LinkSubscription, linkCurrency string) *cf.Subscription {
	out := &cf.Subscription{
		AuthorizationAmountRefund: s.AuthorizationAmountRefund,
	}
	if s.SubscriptionID != "" {
		v := s.SubscriptionID
		out.SubscriptionId = &v
	}
	if s.AuthorizationAmountMinor > 0 {
		amt := float32(currencyutils.AmountMinorToMajor(int64(s.AuthorizationAmountMinor), linkCurrency))
		out.AuthorizationAmount = &amt
	}
	if s.ExpiryTime != nil {
		v := s.ExpiryTime.Format("2006-01-02T15:04:05-07:00")
		out.SubscriptionExpiryTime = &v
	}
	if s.FirstChargeTime != nil {
		v := s.FirstChargeTime.Format("2006-01-02T15:04:05-07:00")
		out.SubscriptionFirstChargeTime = &v
	}
	if s.PlanDetails != nil {
		out.PlanDetails = mapLinkPlanDetails(s.PlanDetails)
	}
	return out
}

// mapLinkPlanDetails maps the reused canonical CreatePlanRequest to cf.CreateLinkPlanRequest. Plan
// amounts are converted minor→major via the plan's own currency (CreatePlanRequest.Currency is
// required). PERIODIC-only fields (amount/intervals/interval_type) are set only when present.
func mapLinkPlanDetails(p *domain.CreatePlanRequest) *cf.CreateLinkPlanRequest {
	planCurrency := string(p.Currency)
	out := &cf.CreateLinkPlanRequest{
		PlanId:        p.PlanID,
		PlanName:      p.PlanName,
		PlanType:      string(p.PlanType),
		PlanMaxAmount: float32(currencyutils.AmountMinorToMajor(int64(p.MaxAmountMinor), planCurrency)),
	}
	if p.Currency != "" {
		out.PlanCurrency = &planCurrency
	}
	if p.AmountMinor > 0 {
		amt := float32(currencyutils.AmountMinorToMajor(int64(p.AmountMinor), planCurrency))
		out.PlanAmount = &amt
	}
	if p.MaxCycles > 0 {
		v := p.MaxCycles
		out.PlanMaxCycles = &v
	}
	if p.Interval > 0 {
		v := p.Interval
		out.PlanIntervals = &v
	}
	if p.IntervalType != "" {
		v := string(p.IntervalType)
		out.PlanIntervalType = &v
	}
	if p.Note != "" {
		v := p.Note
		out.PlanNote = &v
	}
	return out
}
