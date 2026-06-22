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
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment %s not found: %w", req.PaymentID, domain.ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	// Map Razorpay response to canonical domain type
	payment := &domain.Payment{
		ProviderPaymentID: getString(responseMap, "id"),
		OrderID:           getString(responseMap, "order_id"),
		AmountMinor:       domain.AmountMinor(getInt64(responseMap, "amount")),
		Currency:          domain.Currency(getString(responseMap, "currency")),
		Status:            mapPaymentStatus(getString(responseMap, "status")),
		PaymentMethod:     getString(responseMap, "method"),
		IsCaptured:        getBool(responseMap, "captured"),
		BankReference:     getString(responseMap, "bank_account"),
		ErrorCode:         getString(responseMap, "error_code"),
		ErrorMessage:      getString(responseMap, "error_description"),
		PaymentTime:       getTime(responseMap, "created_at"),
		Raw:               rawMapResponse(responseMap),
	}

	return payment, nil
}

// ListPayments retrieves all payments for an order.
// It takes a ListPaymentsRequest containing the order ID and returns a slice of canonical Payment domain objects.
func (a *Adapter) ListPayments(ctx context.Context, req *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
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
		if isNotFoundError(err) {
			return nil, fmt.Errorf("order %s not found: %w", req.OrderID, domain.ErrOrderNotFound)
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
	payments := make([]*domain.Payment, 0, len(itemsList))
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		payment := &domain.Payment{
			ProviderPaymentID: getString(itemMap, "id"),
			OrderID:           getString(itemMap, "order_id"),
			AmountMinor:       domain.AmountMinor(getInt64(itemMap, "amount")),
			Currency:          domain.Currency(getString(itemMap, "currency")),
			Status:            mapPaymentStatus(getString(itemMap, "status")),
			PaymentMethod:     getString(itemMap, "method"),
			IsCaptured:        getBool(itemMap, "captured"),
			BankReference:     getString(itemMap, "bank_account"),
			ErrorCode:         getString(itemMap, "error_code"),
			ErrorMessage:      getString(itemMap, "error_description"),
			PaymentTime:       getTime(itemMap, "created_at"),
			Raw:               rawMapResponse(itemMap),
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
	responseMap, err := a.client.Payment.Capture(req.PaymentID, int(req.AmountMinor), nil, nil)
	if err != nil {
		// Check if payment not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment %s not found: %w", req.PaymentID, domain.ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("failed to capture payment: %w", err)
	}

	// Map Razorpay response to canonical domain type
	payment := &domain.Payment{
		ProviderPaymentID: getString(responseMap, "id"),
		OrderID:           getString(responseMap, "order_id"),
		AmountMinor:       domain.AmountMinor(getInt64(responseMap, "amount")),
		Currency:          domain.Currency(getString(responseMap, "currency")),
		Status:            mapPaymentStatus(getString(responseMap, "status")),
		PaymentMethod:     getString(responseMap, "method"),
		IsCaptured:        getBool(responseMap, "captured"),
		BankReference:     getString(responseMap, "bank_account"),
		ErrorCode:         getString(responseMap, "error_code"),
		ErrorMessage:      getString(responseMap, "error_description"),
		PaymentTime:       getTime(responseMap, "created_at"),
		Raw:               rawMapResponse(responseMap),
	}

	return payment, nil
}
