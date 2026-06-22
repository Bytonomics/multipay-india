package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree_pg"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// getPayment retrieves a specific payment for an order from the Cashfree payment gateway.
// Maps the canonical domain.GetPaymentRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.Payment.
func getPayment(ctx context.Context, adapter *Adapter, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	if req.PaymentID == "" {
		return nil, fmt.Errorf("PaymentID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch payment
	apiVersion := "2022-09-01"
	cfPayment, _, err := cf.PGOrderFetchPaymentWithContext(
		ctx,
		stringPtr(apiVersion),
		req.OrderID,
		req.PaymentID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 payment not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment %s not found: %w", req.PaymentID, domain.ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("failed to fetch payment from Cashfree: %w", domain.ErrProviderError)
	}

	if cfPayment == nil {
		return nil, fmt.Errorf("cashfree returned nil payment: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	payment := MapPaymentEntityToCanonical(cfPayment)
	return payment, nil
}

// listPayments retrieves all payments for a specific order from the Cashfree payment gateway.
// Calls the Cashfree SDK to fetch payments and maps them to canonical domain.Payment types.
func listPayments(ctx context.Context, adapter *Adapter, req *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch payments for the order
	apiVersion := "2022-09-01"
	cfPayments, _, err := cf.PGOrderFetchPaymentsWithContext(
		ctx,
		stringPtr(apiVersion),
		req.OrderID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 order not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("order %s not found: %w", req.OrderID, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("failed to fetch payments from Cashfree: %w", domain.ErrProviderError)
	}

	if cfPayments == nil {
		return []*domain.Payment{}, nil
	}

	// Map response to canonical types
	result := make([]*domain.Payment, 0, len(cfPayments))
	for i := range cfPayments {
		cfPayment := &cfPayments[i]
		payment := MapPaymentEntityToCanonical(cfPayment)
		if payment != nil {
			result = append(result, payment)
		}
	}

	return result, nil
}
