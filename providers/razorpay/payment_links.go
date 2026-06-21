package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// CreatePaymentLink creates a new shareable payment link.
// It takes a CreatePaymentLinkRequest and returns a canonical PaymentLinkResponse domain object.
func (a *Adapter) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.Amount <= 0 {
		return nil, domain.ErrInvalidRequest
	}
	if req.Currency == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Build Razorpay payment link creation parameters
	params := map[string]interface{}{
		"amount":   req.Amount,
		"currency": req.Currency,
	}

	// Add optional fields if provided
	if req.Description != "" {
		params["description"] = req.Description
	}

	if len(req.Notes) > 0 {
		params["notes"] = req.Notes
	}

	if req.NotifyEmail != "" {
		params["notify"] = map[string]interface{}{
			"email": true,
		}
		params["notify_email"] = req.NotifyEmail
	}

	if req.NotifyPhone != "" {
		params["notify"] = map[string]interface{}{
			"sms": true,
		}
		params["notify_phone"] = req.NotifyPhone
	}

	if req.ExpiresAt != nil {
		params["expire_by"] = req.ExpiresAt.Unix()
	}

	// Call Razorpay SDK to create payment link
	// Razorpay PaymentLink.Create signature: Create(params map[string]interface{}, options map[string]string)
	responseMap, err := a.client.PaymentLink.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	// Map Razorpay response to canonical domain type
	linkResponse := &domain.PaymentLinkResponse{
		ID:          getString(responseMap, "id"),
		URL:         getString(responseMap, "short_url"),
		Amount:      getInt64(responseMap, "amount"),
		Currency:    getString(responseMap, "currency"),
		Description: getString(responseMap, "description"),
		Status:      getString(responseMap, "status"),
		NotifyEmail: getString(responseMap, "notify_email"),
		NotifyPhone: getString(responseMap, "notify_phone"),
		CreatedAt:   getTime(responseMap, "created_at"),
		UpdatedAt:   getTime(responseMap, "updated_at"),
	}

	// Handle optional ExpiresAt field
	if expireBy, ok := responseMap["expire_by"].(float64); ok && expireBy > 0 {
		expireTime := getTime(responseMap, "expire_by")
		linkResponse.ExpiresAt = &expireTime
	}

	return linkResponse, nil
}

// GetPaymentLink retrieves an existing payment link.
// It takes a GetPaymentLinkRequest with link ID and returns a canonical PaymentLinkResponse domain object.
func (a *Adapter) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
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
		if err.Error() == "Payment link not found" {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, fmt.Errorf("failed to fetch payment link: %w", err)
	}

	// Map Razorpay response to canonical domain type
	linkResponse := &domain.PaymentLinkResponse{
		ID:          getString(responseMap, "id"),
		URL:         getString(responseMap, "short_url"),
		Amount:      getInt64(responseMap, "amount"),
		Currency:    getString(responseMap, "currency"),
		Description: getString(responseMap, "description"),
		Status:      getString(responseMap, "status"),
		NotifyEmail: getString(responseMap, "notify_email"),
		NotifyPhone: getString(responseMap, "notify_phone"),
		CreatedAt:   getTime(responseMap, "created_at"),
		UpdatedAt:   getTime(responseMap, "updated_at"),
	}

	// Handle optional ExpiresAt field
	if expireBy, ok := responseMap["expire_by"].(float64); ok && expireBy > 0 {
		expireTime := getTime(responseMap, "expire_by")
		linkResponse.ExpiresAt = &expireTime
	}

	return linkResponse, nil
}

// CancelPaymentLink cancels an existing payment link.
// It takes a CancelPaymentLinkRequest with link ID and returns the updated PaymentLinkResponse domain object.
func (a *Adapter) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
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
		if err.Error() == "Payment link not found" {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, fmt.Errorf("failed to cancel payment link: %w", err)
	}

	// Map Razorpay response to canonical domain type
	linkResponse := &domain.PaymentLinkResponse{
		ID:          getString(responseMap, "id"),
		URL:         getString(responseMap, "short_url"),
		Amount:      getInt64(responseMap, "amount"),
		Currency:    getString(responseMap, "currency"),
		Description: getString(responseMap, "description"),
		Status:      getString(responseMap, "status"),
		NotifyEmail: getString(responseMap, "notify_email"),
		NotifyPhone: getString(responseMap, "notify_phone"),
		CreatedAt:   getTime(responseMap, "created_at"),
		UpdatedAt:   getTime(responseMap, "updated_at"),
	}

	// Handle optional ExpiresAt field
	if expireBy, ok := responseMap["expire_by"].(float64); ok && expireBy > 0 {
		expireTime := getTime(responseMap, "expire_by")
		linkResponse.ExpiresAt = &expireTime
	}

	return linkResponse, nil
}
