package razorpay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// D17: Typed request structs for Razorpay Payment Link APIs.
type razorpayNotifyOptions struct {
	Email bool `json:"email,omitempty"`
	SMS   bool `json:"sms,omitempty"`
}

type razorpayCreatePaymentLinkRequest struct {
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description,omitempty"`
	Notes       domain.Metadata        `json:"notes,omitempty"`
	Notify      *razorpayNotifyOptions `json:"notify,omitempty"`
	ExpireBy    int64                  `json:"expire_by,omitempty"`
}

// D17: Typed response struct for Razorpay Payment Link API responses.
type razorpayPaymentLinkResponse struct {
	ID                    string `json:"id"`
	Entity                string `json:"entity"`
	Description           string `json:"description"`
	Amount                int64  `json:"amount"`
	AmountPaid            int64  `json:"amount_paid"`
	Currency              string `json:"currency"`
	Status                string `json:"status"`
	ShortURL              string `json:"short_url"`
	CallbackURL           string `json:"callback_url"`
	CallbackMethod        string `json:"callback_method"`
	ReminderEnable        bool   `json:"reminder_enable"`
	PaymentsCount         int64  `json:"payments_count"`
	FirstMinPartialAmount int64  `json:"first_min_partial_amount"`
	CreatedAt             int64  `json:"created_at"`
	ExpireBy              int64  `json:"expire_by"`
}

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
	linkReq := &razorpayCreatePaymentLinkRequest{
		Amount:      int64(req.AmountMinor),
		Currency:    string(req.Currency),
		Description: req.Purpose,
		Notes:       req.Metadata,
	}
	if req.NotifyEmail != nil && *req.NotifyEmail {
		if linkReq.Notify == nil {
			linkReq.Notify = &razorpayNotifyOptions{}
		}
		linkReq.Notify.Email = true
	}
	if req.NotifySMS != nil && *req.NotifySMS {
		if linkReq.Notify == nil {
			linkReq.Notify = &razorpayNotifyOptions{}
		}
		linkReq.Notify.SMS = true
	}
	if req.ExpiryTime != nil {
		linkReq.ExpireBy = req.ExpiryTime.Unix()
	}
	params, err := encodeRequest(linkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode create payment link request: %w", err)
	}

	// Call Razorpay SDK to create payment link
	// Razorpay PaymentLink.Create signature: Create(params map[string]interface{}, options map[string]string)
	responseMap, err := a.client.PaymentLink.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentLinkResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment link response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapPaymentLinkFromResponse(typed, rawJSON), nil
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

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentLinkResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment link response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapPaymentLinkFromResponse(typed, rawJSON), nil
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

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentLinkResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment link response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapPaymentLinkFromResponse(typed, rawJSON), nil
}

// D17: Typed struct mapper for payment link response
func mapPaymentLinkFromResponse(r *razorpayPaymentLinkResponse, rawJSON []byte) *domain.PaymentLink {
	linkResponse := &domain.PaymentLink{
		ProviderLinkID: r.ID,
		LinkID:         r.ID,
		LinkURL:        r.ShortURL,
		AmountMinor:    domain.AmountMinor(r.Amount),
		AmountPaid:     domain.AmountMinor(r.AmountPaid),
		Currency:       domain.Currency(r.Currency),
		Purpose:        r.Description,
		Status:         domain.PaymentLinkStatus(r.Status),
		CreatedAt:      unixPtr(r.CreatedAt),
		Raw:            domain.RawProviderResponse(rawJSON),
		ProviderDetails: &domain.PaymentLinkProviderDetail{
			Razorpay: &domain.RazorpayPaymentLinkDetail{
				Entity:          r.Entity,
				Description:     r.Description,
				CallbackURL:     r.CallbackURL,
				CallbackMethod:  r.CallbackMethod,
				ReminderEnable:  r.ReminderEnable,
				PaymentsCount:   r.PaymentsCount,
				FirstMinPartial: r.FirstMinPartialAmount,
			},
		},
	}

	// D17: Handle optional ExpiryTime field
	if r.ExpireBy > 0 {
		linkResponse.ExpiryTime = unixPtr(r.ExpireBy)
	}

	return linkResponse
}
