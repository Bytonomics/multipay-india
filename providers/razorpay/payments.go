package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// GetPayment retrieves a specific payment for an order.
// It takes a GetPaymentRequest containing order and payment IDs and returns a canonical Payment domain object.
func (a *Adapter) GetPayment(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.PaymentID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch payment
	responseMap, err := a.client.Payment.Fetch(req.PaymentID, nil, nil)
	if err != nil {
		// Check if payment not found
		if err.Error() == "Payment not found" {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	// Map Razorpay response to canonical domain type
	payment := &domain.Payment{
		ID:        getString(responseMap, "id"),
		OrderID:   getString(responseMap, "order_id"),
		Amount:    getInt64(responseMap, "amount"),
		Status:    mapPaymentStatus(getString(responseMap, "status")),
		Method:    getString(responseMap, "method"),
		CreatedAt: getTime(responseMap, "created_at"),
	}

	return payment, nil
}

// ListPayments retrieves all payments for an order.
// It takes a GetOrderRequest containing the order ID and returns a slice of canonical Payment domain objects.
func (a *Adapter) ListPayments(ctx context.Context, req *domain.GetOrderRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Build parameters to filter payments by order
	params := map[string]interface{}{
		"order_id": req.OrderID,
	}

	// Call Razorpay SDK to fetch payments
	paymentsData, err := a.client.Payment.All(params, nil)
	if err != nil {
		// Check if order not found
		if err.Error() == "Order not found" {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	// Handle the response - Razorpay returns a map with "items" key containing payment list
	itemsList, ok := paymentsData["items"].([]interface{})
	if !ok {
		// No items found, return empty slice
		return []*domain.Payment{}, nil
	}

	// Map each payment response to canonical domain type
	var payments []*domain.Payment
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		payment := &domain.Payment{
			ID:        getString(itemMap, "id"),
			OrderID:   getString(itemMap, "order_id"),
			Amount:    getInt64(itemMap, "amount"),
			Status:    mapPaymentStatus(getString(itemMap, "status")),
			Method:    getString(itemMap, "method"),
			CreatedAt: getTime(itemMap, "created_at"),
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// CapturePayment captures an authorized payment (Razorpay-specific capability).
// It takes a CapturePaymentRequest with payment ID and amount, and returns the captured Payment.
// This method is specific to Razorpay and may not be available on all providers.
func (a *Adapter) CapturePayment(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.PaymentID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to capture payment
	// Razorpay Capture method signature: Capture(paymentID string, amount int, options, headers map[string]string)
	responseMap, err := a.client.Payment.Capture(req.PaymentID, int(req.Amount), nil, nil)
	if err != nil {
		// Check if payment not found
		if err.Error() == "Payment not found" {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to capture payment: %w", err)
	}

	// Map Razorpay response to canonical domain type
	payment := &domain.Payment{
		ID:        getString(responseMap, "id"),
		OrderID:   getString(responseMap, "order_id"),
		Amount:    getInt64(responseMap, "amount"),
		Status:    mapPaymentStatus(getString(responseMap, "status")),
		Method:    getString(responseMap, "method"),
		CreatedAt: getTime(responseMap, "created_at"),
	}

	return payment, nil
}
