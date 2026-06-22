package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// CreatePaymentLink creates a new shareable payment link.
// It takes a CreatePaymentLinkRequest and returns a canonical PaymentLinkResponse domain object.
func (a *Adapter) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.AmountMinor <= 0 {
		return nil, domain.ErrInvalidRequest
	}
	if req.Currency == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Build Razorpay payment link creation parameters
	params := map[string]interface{}{
		"amount":   req.AmountMinor,
		"currency": string(req.Currency),
	}

	// Add optional fields if provided
	if req.Purpose != "" {
		params["description"] = req.Purpose
	}

	if len(req.Metadata) > 0 {
		params["notes"] = req.Metadata
	}

	if req.NotifyEmail != nil && *req.NotifyEmail {
		params["notify"] = map[string]interface{}{
			"email": true,
		}
	}

	if req.NotifySMS != nil && *req.NotifySMS {
		params["notify"] = map[string]interface{}{
			"sms": true,
		}
	}

	if req.ExpiryTime != nil {
		params["expire_by"] = req.ExpiryTime.Unix()
	}

	// Call Razorpay SDK to create payment link
	// Razorpay PaymentLink.Create signature: Create(params map[string]interface{}, options map[string]string)
	responseMap, err := a.client.PaymentLink.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	// Map Razorpay response to canonical domain type
	linkResponse := &domain.PaymentLink{
		ProviderLinkID: getString(responseMap, "id"),
		LinkID:         getString(responseMap, "id"),
		LinkURL:        getString(responseMap, "short_url"),
		AmountMinor:    domain.AmountMinor(getInt64(responseMap, "amount")),
		AmountPaid:     domain.AmountMinor(getInt64(responseMap, "amount_paid")),
		Currency:       domain.Currency(getString(responseMap, "currency")),
		Purpose:        getString(responseMap, "description"),
		Status:         domain.PaymentLinkStatus(getString(responseMap, "status")),
		CreatedAt:      getTime(responseMap, "created_at"),
		Raw:            rawMapResponse(responseMap),
		ProviderDetails: &domain.PaymentLinkProviderDetail{
			Razorpay: &domain.RazorpayPaymentLinkDetail{
				Entity:          getString(responseMap, "entity"),
				Description:     getString(responseMap, "description"),
				CallbackURL:     getString(responseMap, "callback_url"),
				CallbackMethod:  getString(responseMap, "callback_method"),
				ReminderEnable:  getBool(responseMap, "reminder_enable"),
				PaymentsCount:   getInt64(responseMap, "payments_count"),
				FirstMinPartial: getInt64(responseMap, "first_min_partial_amount"),
			},
		},
	}

	// Handle optional ExpiryTime field
	if expireBy, ok := responseMap["expire_by"].(float64); ok && expireBy > 0 {
		linkResponse.ExpiryTime = getTime(responseMap, "expire_by")
	}

	return linkResponse, nil
}

// GetPaymentLink retrieves an existing payment link.
// It takes a GetPaymentLinkRequest with link ID and returns a canonical PaymentLinkResponse domain object.
func (a *Adapter) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.LinkID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch payment link
	// Razorpay PaymentLink.Fetch signature: Fetch(linkID string, options map[string]interface{}, headers map[string]string)
	responseMap, err := a.client.PaymentLink.Fetch(req.LinkID, nil, nil)
	if err != nil {
		// Check if payment link not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment link %s not found: %w", req.LinkID, domain.ErrPaymentLinkNotFound)
		}
		return nil, fmt.Errorf("failed to fetch payment link: %w", err)
	}

	// Map Razorpay response to canonical domain type
	linkResponse := &domain.PaymentLink{
		ProviderLinkID: getString(responseMap, "id"),
		LinkID:         getString(responseMap, "id"),
		LinkURL:        getString(responseMap, "short_url"),
		AmountMinor:    domain.AmountMinor(getInt64(responseMap, "amount")),
		AmountPaid:     domain.AmountMinor(getInt64(responseMap, "amount_paid")),
		Currency:       domain.Currency(getString(responseMap, "currency")),
		Purpose:        getString(responseMap, "description"),
		Status:         domain.PaymentLinkStatus(getString(responseMap, "status")),
		CreatedAt:      getTime(responseMap, "created_at"),
		Raw:            rawMapResponse(responseMap),
		ProviderDetails: &domain.PaymentLinkProviderDetail{
			Razorpay: &domain.RazorpayPaymentLinkDetail{
				Entity:          getString(responseMap, "entity"),
				Description:     getString(responseMap, "description"),
				CallbackURL:     getString(responseMap, "callback_url"),
				CallbackMethod:  getString(responseMap, "callback_method"),
				ReminderEnable:  getBool(responseMap, "reminder_enable"),
				PaymentsCount:   getInt64(responseMap, "payments_count"),
				FirstMinPartial: getInt64(responseMap, "first_min_partial_amount"),
			},
		},
	}

	// Handle optional ExpiryTime field
	if expireBy, ok := responseMap["expire_by"].(float64); ok && expireBy > 0 {
		linkResponse.ExpiryTime = getTime(responseMap, "expire_by")
	}

	return linkResponse, nil
}

// CancelPaymentLink cancels an existing payment link.
// It takes a CancelPaymentLinkRequest with link ID and returns the updated PaymentLinkResponse domain object.
func (a *Adapter) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.LinkID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to cancel payment link
	// Razorpay PaymentLink.Cancel signature: Cancel(linkID string, options map[string]interface{}, headers map[string]string)
	responseMap, err := a.client.PaymentLink.Cancel(req.LinkID, nil, nil)
	if err != nil {
		// Check if payment link not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment link %s not found: %w", req.LinkID, domain.ErrPaymentLinkNotFound)
		}
		return nil, fmt.Errorf("failed to cancel payment link: %w", err)
	}

	// Map Razorpay response to canonical domain type
	linkResponse := &domain.PaymentLink{
		ProviderLinkID: getString(responseMap, "id"),
		LinkID:         getString(responseMap, "id"),
		LinkURL:        getString(responseMap, "short_url"),
		AmountMinor:    domain.AmountMinor(getInt64(responseMap, "amount")),
		AmountPaid:     domain.AmountMinor(getInt64(responseMap, "amount_paid")),
		Currency:       domain.Currency(getString(responseMap, "currency")),
		Purpose:        getString(responseMap, "description"),
		Status:         domain.PaymentLinkStatus(getString(responseMap, "status")),
		CreatedAt:      getTime(responseMap, "created_at"),
		Raw:            rawMapResponse(responseMap),
		ProviderDetails: &domain.PaymentLinkProviderDetail{
			Razorpay: &domain.RazorpayPaymentLinkDetail{
				Entity:          getString(responseMap, "entity"),
				Description:     getString(responseMap, "description"),
				CallbackURL:     getString(responseMap, "callback_url"),
				CallbackMethod:  getString(responseMap, "callback_method"),
				ReminderEnable:  getBool(responseMap, "reminder_enable"),
				PaymentsCount:   getInt64(responseMap, "payments_count"),
				FirstMinPartial: getInt64(responseMap, "first_min_partial_amount"),
			},
		},
	}

	// Handle optional ExpiryTime field
	if expireBy, ok := responseMap["expire_by"].(float64); ok && expireBy > 0 {
		linkResponse.ExpiryTime = getTime(responseMap, "expire_by")
	}

	return linkResponse, nil
}
