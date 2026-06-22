package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree_pg"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// getInstrument retrieves a specific payment instrument from the Cashfree payment gateway.
// Maps the canonical domain.GetInstrumentRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.Instrument.
func getInstrument(ctx context.Context, adapter *Adapter, req *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.CustomerID == "" {
		return nil, fmt.Errorf("CustomerID is required: %w", domain.ErrInvalidRequest)
	}

	if req.InstrumentID == "" {
		return nil, fmt.Errorf("InstrumentID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch instrument
	apiVersion := "2022-09-01"
	cfInstrument, _, err := cf.PGCustomerFetchInstrumentWithContext(
		ctx,
		stringPtr(apiVersion),
		req.CustomerID,
		req.InstrumentID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 instrument not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("instrument %s not found: %w", req.InstrumentID, domain.ErrInstrumentNotFound)
		}
		return nil, fmt.Errorf("failed to fetch instrument from Cashfree: %w", domain.ErrProviderError)
	}

	if cfInstrument == nil {
		return nil, fmt.Errorf("cashfree returned nil instrument: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	instrument := MapInstrumentEntityToCanonical(cfInstrument)
	return instrument, nil
}

// listInstruments retrieves all payment instruments for a customer from the Cashfree payment gateway.
// Calls the Cashfree SDK to fetch instruments and maps them to canonical domain.Instrument types.
func listInstruments(ctx context.Context, adapter *Adapter, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.CustomerID == "" {
		return nil, fmt.Errorf("CustomerID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch instruments for the customer
	apiVersion := "2022-09-01"
	cfInstruments, _, err := cf.PGCustomerFetchInstrumentsWithContext(
		ctx,
		stringPtr(apiVersion),
		req.CustomerID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // instrumentType (optional filter)
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 customer not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("customer %s not found: %w", req.CustomerID, domain.ErrInstrumentNotFound)
		}
		return nil, fmt.Errorf("failed to fetch instruments from Cashfree: %w", domain.ErrProviderError)
	}

	if cfInstruments == nil {
		return []*domain.Instrument{}, nil
	}

	// Map response to canonical types
	result := make([]*domain.Instrument, 0, len(cfInstruments))
	for i := range cfInstruments {
		cfInstrument := &cfInstruments[i]
		instrument := MapInstrumentEntityToCanonical(cfInstrument)
		if instrument != nil {
			result = append(result, instrument)
		}
	}

	return result, nil
}

// deleteInstrument deletes a payment instrument from the Cashfree payment gateway.
// Returns the deleted instrument on success.
func deleteInstrument(ctx context.Context, adapter *Adapter, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.CustomerID == "" {
		return nil, fmt.Errorf("CustomerID is required: %w", domain.ErrInvalidRequest)
	}

	if req.InstrumentID == "" {
		return nil, fmt.Errorf("InstrumentID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to delete instrument
	apiVersion := "2022-09-01"
	_, _, err := cf.PGCustomerDeleteInstrumentWithContext(
		ctx,
		stringPtr(apiVersion),
		req.CustomerID,
		req.InstrumentID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 instrument not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("instrument %s not found: %w", req.InstrumentID, domain.ErrInstrumentNotFound)
		}
		return nil, fmt.Errorf("failed to delete instrument on Cashfree: %w", domain.ErrProviderError)
	}

	// Return the deleted instrument (constructed from request data)
	instrument := &domain.Instrument{
		InstrumentID: req.InstrumentID,
		CustomerID:   req.CustomerID,
	}

	return instrument, nil
}
