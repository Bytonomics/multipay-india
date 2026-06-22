package cashfree

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cf "github.com/cashfree/cashfree_pg"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// createOrder creates a new order on the Cashfree payment gateway.
// Maps the canonical domain.CreateOrderRequest to a Cashfree CreateOrderRequest,
// calls the SDK, and maps the response back to a canonical domain.Order.
func createOrder(ctx context.Context, adapter *Adapter, req *domain.CreateOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.AmountMinor <= 0 {
		return nil, fmt.Errorf("amount must be positive: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Build Cashfree CreateOrderRequest
	cfReq := &cf.CreateOrderRequest{
		OrderAmount:   AmountMinorToCashfree(int64(req.AmountMinor)),
		OrderCurrency: string(req.Currency),
		CustomerDetails: cf.CustomerDetails{
			CustomerId:    req.Customer.CustomerID,
			CustomerEmail: *cf.NewNullableString(stringPtr(req.Customer.Email)),
			CustomerPhone: req.Customer.Phone,
		},
	}

	// Call Cashfree SDK
	apiVersion := "2022-09-01"
	cfOrder, _, err := cf.PGCreateOrderWithContext(
		ctx,
		stringPtr(apiVersion),
		cfReq,
		nil, // xRequestId
		nil, // xIdempotencyKey
		nil, // httpClient (uses default)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order on cashfree: %w", domain.ErrProviderError)
	}

	if cfOrder == nil {
		return nil, fmt.Errorf("cashfree returned nil order: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	order := MapOrderEntityToCanonical(cfOrder)
	return order, nil
}

// getOrder retrieves an existing order from the Cashfree payment gateway.
// Maps the canonical domain.GetOrderRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.Order.
func getOrder(ctx context.Context, adapter *Adapter, req *domain.GetOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	// Lock the Cashfree SDK and set up globals
	adapter.lockCashfreeSDK()
	defer adapter.unlockCashfreeSDK()

	// Call Cashfree SDK to fetch order
	apiVersion := "2022-09-01"
	cfOrder, _, err := cf.PGFetchOrderWithContext(
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
		return nil, fmt.Errorf("failed to fetch order from Cashfree: %w", domain.ErrProviderError)
	}

	if cfOrder == nil {
		return nil, fmt.Errorf("cashfree returned nil order: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	order := MapOrderEntityToCanonical(cfOrder)
	return order, nil
}

// listOrderPayments retrieves all payments associated with a specific order.
// Calls the Cashfree SDK to fetch payments and maps them to canonical domain.Payment types.
func listOrderPayments(ctx context.Context, adapter *Adapter, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
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

// stringPtr is a helper to create a pointer to a string.
func stringPtr(s string) *string {
	return &s
}

// isNotFoundError checks if an error from Cashfree SDK indicates a 404 not found response.
func isNotFoundError(err error) bool {
	var genericErr cf.GenericOpenAPIError
	if errors.As(err, &genericErr) {
		// Check HTTP status code or error message for 404 indication
		// Cashfree returns GenericOpenAPIError with status code in the error string
		return strings.Contains(strings.ToLower(genericErr.Error()), "404") ||
			strings.Contains(strings.ToLower(genericErr.Error()), "not found")
	}
	return false
}
