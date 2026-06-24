package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// D17: Typed response struct for Razorpay Instrument API responses.
type razorpayInstrumentResponse struct {
	ID               string `json:"id"`
	Entity           string `json:"entity"`
	CustomerID       string `json:"customer_id"`
	Method           string `json:"method"`
	DisplayValue     string `json:"display_value"`
	Status           string `json:"status"`
	CreatedAt        int64  `json:"created_at"`
	Token            string `json:"token"`
	MaxPaymentAmount int64  `json:"max_payment_amount"`
	ExpiredAt        int64  `json:"expired_at"`
	Compliant        bool   `json:"compliant"`
}

// D17: Typed response struct for instrument list response.
type razorpayInstrumentListResponse struct {
	Items []razorpayInstrumentResponse `json:"items"`
}

// GetInstrument retrieves a specific payment instrument (called "token" in Razorpay).
// It takes a GetInstrumentRequest with customer and instrument IDs and returns a canonical Instrument domain object.
func (a *Adapter) GetInstrument(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.CustomerID == "" {
		return nil, domain.ErrInvalidRequest
	}
	if req.InstrumentID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch token (instrument)
	// Razorpay Token.Fetch signature: Fetch(customerID string, tokenID string, options map[string]any, headers map[string]string)
	responseMap, err := a.client.Token.Fetch(req.CustomerID, req.InstrumentID, make(map[string]any), make(map[string]string))
	if err != nil {
		// Check if token not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("instrument %s not found: %w", req.InstrumentID, domain.ErrInstrumentNotFound)
		}
		return nil, fmt.Errorf("failed to fetch instrument: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayInstrumentResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Map typed struct to canonical domain type
	return mapInstrumentFromResponse(typed, responseMap), nil
}

// ListInstruments retrieves all instruments for a customer (called "tokens" in Razorpay).
// It takes a ListInstrumentsRequest with customer ID and returns a slice of canonical Instrument domain objects.
func (a *Adapter) ListInstruments(ctx context.Context, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.CustomerID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch tokens (instruments)
	// Razorpay Token.All signature: All(customerID string, params map[string]any, options map[string]string)
	tokensData, err := a.client.Token.All(req.CustomerID, nil, nil)
	if err != nil {
		// Check if customer not found
		if isNotFoundError(err) {
			return nil, domain.ErrInvalidRequest
		}
		return nil, fmt.Errorf("failed to list instruments: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayInstrumentListResponse](tokensData)
	if err != nil {
		return nil, err
	}

	// D17: Map each typed instrument response to canonical domain type
	instruments := make([]*domain.Instrument, 0, len(typed.Items))
	for i := range typed.Items {
		instrument := mapInstrumentFromResponse(&typed.Items[i], tokensData)
		instruments = append(instruments, instrument)
	}

	return instruments, nil
}

// DeleteInstrument removes a payment instrument (called "token" in Razorpay).
// It takes a DeleteInstrumentRequest with customer and instrument IDs and returns the deleted Instrument domain object.
func (a *Adapter) DeleteInstrument(ctx context.Context, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.CustomerID == "" {
		return nil, domain.ErrInvalidRequest
	}
	if req.InstrumentID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to delete token (instrument)
	// Razorpay Token.Delete signature: Delete(customerID string, tokenID string, options map[string]any, headers map[string]string)
	responseMap, err := a.client.Token.Delete(req.CustomerID, req.InstrumentID, make(map[string]any), make(map[string]string))
	if err != nil {
		// Check if token not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("instrument %s not found: %w", req.InstrumentID, domain.ErrInstrumentNotFound)
		}
		return nil, fmt.Errorf("failed to delete instrument: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayInstrumentResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Map typed struct to canonical domain type
	return mapInstrumentFromResponse(typed, responseMap), nil
}

// D17: Typed struct mapper for instrument response
func mapInstrumentFromResponse(r *razorpayInstrumentResponse, raw map[string]any) *domain.Instrument {
	return &domain.Instrument{
		InstrumentID:   r.ID,
		CustomerID:     r.CustomerID,
		InstrumentType: r.Method, // Razorpay uses "method" field for instrument type
		DisplayValue:   r.DisplayValue,
		Status:         r.Status,
		CreatedAt:      unixPtr(r.CreatedAt),
		Raw:            rawMapResponse(raw),
		ProviderDetails: &domain.InstrumentProviderDetail{
			Razorpay: &domain.RazorpayInstrumentDetail{
				Entity:           r.Entity,
				Token:            r.Token,
				MaxPaymentAmount: r.MaxPaymentAmount,
				ExpiredAt:        r.ExpiredAt,
				Compliant:        r.Compliant,
			},
		},
	}
}
