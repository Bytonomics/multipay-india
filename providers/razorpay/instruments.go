package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

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
	// Razorpay Token.Fetch signature: Fetch(customerID string, tokenID string, options map[string]interface{}, headers map[string]string)
	responseMap, err := a.client.Token.Fetch(req.CustomerID, req.InstrumentID, nil, nil)
	if err != nil {
		// Check if token not found
		if err.Error() == "Token not found" {
			return nil, domain.ErrInstrumentNotFound
		}
		return nil, fmt.Errorf("failed to fetch instrument: %w", err)
	}

	// Map Razorpay response to canonical domain type
	instrument := &domain.Instrument{
		InstrumentID:   getString(responseMap, "id"),
		CustomerID:     getString(responseMap, "customer_id"),
		InstrumentType: getString(responseMap, "method"), // Razorpay uses "method" field for instrument type
		CreatedAt:      getTime(responseMap, "created_at"),
	}

	return instrument, nil
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

	// Build parameters for listing tokens
	params := make(map[string]interface{})

	// Call Razorpay SDK to fetch tokens (instruments)
	// Razorpay Token.All signature: All(customerID string, params map[string]interface{}, options map[string]string)
	tokensData, err := a.client.Token.All(req.CustomerID, params, nil)
	if err != nil {
		// Check if customer not found
		if err.Error() == "Customer not found" {
			return nil, domain.ErrInvalidRequest
		}
		return nil, fmt.Errorf("failed to list instruments: %w", err)
	}

	// Handle the response - Razorpay returns a map with "items" key containing token list
	itemsList, ok := tokensData["items"].([]interface{})
	if !ok {
		// No items found, return empty slice
		return []*domain.Instrument{}, nil
	}

	// Map each token response to canonical domain type
	instruments := make([]*domain.Instrument, 0, len(itemsList))
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		instrument := &domain.Instrument{
			InstrumentID:   getString(itemMap, "id"),
			CustomerID:     getString(itemMap, "customer_id"),
			InstrumentType: getString(itemMap, "method"), // Razorpay uses "method" field for instrument type
			CreatedAt:      getTime(itemMap, "created_at"),
		}

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
	// Razorpay Token.Delete signature: Delete(customerID string, tokenID string, options map[string]interface{}, headers map[string]string)
	responseMap, err := a.client.Token.Delete(req.CustomerID, req.InstrumentID, nil, nil)
	if err != nil {
		// Check if token not found
		if err.Error() == "Token not found" {
			return nil, domain.ErrInstrumentNotFound
		}
		return nil, fmt.Errorf("failed to delete instrument: %w", err)
	}

	// Map Razorpay response to canonical domain type
	instrument := &domain.Instrument{
		InstrumentID:   getString(responseMap, "id"),
		CustomerID:     getString(responseMap, "customer_id"),
		InstrumentType: getString(responseMap, "method"), // Razorpay uses "method" field for instrument type
		CreatedAt:      getTime(responseMap, "created_at"),
	}

	return instrument, nil
}
