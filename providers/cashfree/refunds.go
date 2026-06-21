package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree_pg"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// createRefund creates a new refund for an order on the Cashfree payment gateway.
// Maps the canonical domain.CreateRefundRequest to a Cashfree refund request,
// calls the SDK, and maps the response back to a canonical domain.Refund.
func createRefund(ctx context.Context, adapter *Adapter, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Build Cashfree refund request
	refundAmount := 0.0
	if req.Amount > 0 {
		refundAmount = AmountMinorToCashfree(req.Amount)
	}

	cfReq := &cf.OrderCreateRefundRequest{
		RefundAmount: refundAmount,
		RefundNote:   stringPtr(req.Notes),
	}

	// Call Cashfree SDK to create refund
	apiVersion := "2022-09-01"
	cfRefund, _, err := cf.PGOrderCreateRefund(
		stringPtr(apiVersion),
		req.OrderID,
		cfReq,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund on Cashfree: %w", domain.ErrProviderError)
	}

	if cfRefund == nil {
		return nil, fmt.Errorf("Cashfree returned nil refund: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	refund := MapRefundEntityToCanonical(cfRefund)
	return refund, nil
}

// getRefund retrieves a specific refund from the Cashfree payment gateway.
// Maps the canonical domain.GetRefundRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.Refund.
func getRefund(ctx context.Context, adapter *Adapter, req *domain.GetRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	if req.RefundID == "" {
		return nil, fmt.Errorf("RefundID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch refund
	apiVersion := "2022-09-01"
	cfRefund, _, err := cf.PGOrderFetchRefund(
		stringPtr(apiVersion),
		req.OrderID,
		req.RefundID,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		// Check if error is 404 refund not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("refund %s not found: %w", req.RefundID, domain.ErrRefundNotFound)
		}
		return nil, fmt.Errorf("failed to fetch refund from Cashfree: %w", domain.ErrProviderError)
	}

	if cfRefund == nil {
		return nil, fmt.Errorf("Cashfree returned nil refund: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	refund := MapRefundEntityToCanonical(cfRefund)
	return refund, nil
}

// listRefunds retrieves all refunds for a specific order from the Cashfree payment gateway.
// Calls the Cashfree SDK to fetch refunds and maps them to canonical domain.Refund types.
func listRefunds(ctx context.Context, adapter *Adapter, req *domain.GetOrderRequest) ([]*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch refunds for the order
	apiVersion := "2022-09-01"
	cfRefunds, _, err := cf.PGOrderFetchRefunds(
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
		return nil, fmt.Errorf("failed to fetch refunds from Cashfree: %w", domain.ErrProviderError)
	}

	if cfRefunds == nil {
		return []*domain.Refund{}, nil
	}

	// Map response to canonical types
	result := make([]*domain.Refund, 0)
	for i := range cfRefunds {
		cfRefund := &cfRefunds[i]
		refund := MapRefundEntityToCanonical(cfRefund)
		if refund != nil {
			result = append(result, refund)
		}
	}

	return result, nil
}
