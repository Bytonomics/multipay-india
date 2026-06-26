package cashfree

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
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

	// Call Cashfree SDK to fetch payment
	cfPayment, httpResp, err := adapter.cfClient.PGOrderFetchPaymentWithContext(
		ctx,
		req.OrderID,
		req.PaymentID,
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
	payment, err := MapPaymentEntityToCanonical(cfPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to map payment: %w", err)
	}
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

	// Call Cashfree SDK to fetch payments for the order
	cfPayments, httpResp, err := adapter.cfClient.PGOrderFetchPaymentsWithContext(
		ctx,
		req.OrderID,
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
		payment, err := MapPaymentEntityToCanonical(cfPayment)
		if err != nil {
			return nil, fmt.Errorf("failed to map payment at index %d: %w", i, err)
		}
		if payment != nil {
			result = append(result, payment)
		}
	}

	return result, nil
}
