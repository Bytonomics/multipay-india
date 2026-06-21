package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// CreateRefund creates a new refund for a payment.
// It takes a CreateRefundRequest with payment ID and optional amount, and returns a canonical Refund domain object.
func (a *Adapter) CreateRefund(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Build Razorpay refund creation parameters
	params := make(map[string]interface{})

	// Payment ID is required - In Razorpay, refunds are created against a payment ID
	// req.OrderID is used as the payment ID; caller must provide correct ID
	params["payment_id"] = req.OrderID

	// Amount is optional for full refund; if provided, partial refund
	if req.Amount > 0 {
		params["amount"] = req.Amount
	}

	// Add notes if provided
	if req.Notes != "" {
		params["notes"] = map[string]string{
			"note": req.Notes,
		}
	}

	// Call Razorpay SDK to create refund
	// Razorpay Create refund method signature: Create(params map[string]interface{}, options map[string]string)
	responseMap, err := a.client.Refund.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Map Razorpay response to canonical domain type
	refund := &domain.Refund{
		ID:        getString(responseMap, "id"),
		OrderID:   getString(responseMap, "payment_id"), // Razorpay returns payment_id, map to OrderID for domain
		Amount:    getInt64(responseMap, "amount"),
		Status:    mapRefundStatus(getString(responseMap, "status")),
		CreatedAt: getTime(responseMap, "created_at"),
	}

	return refund, nil
}

// GetRefund retrieves an existing refund.
// It takes a GetRefundRequest with refund ID and returns a canonical Refund domain object.
func (a *Adapter) GetRefund(ctx context.Context, req *domain.GetRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.RefundID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch refund
	responseMap, err := a.client.Refund.Fetch(req.RefundID, nil, nil)
	if err != nil {
		// Check if refund not found
		if err.Error() == "Refund not found" {
			return nil, domain.ErrRefundNotFound
		}
		return nil, fmt.Errorf("failed to fetch refund: %w", err)
	}

	// Map Razorpay response to canonical domain type
	refund := &domain.Refund{
		ID:        getString(responseMap, "id"),
		OrderID:   getString(responseMap, "payment_id"), // Razorpay returns payment_id, map to OrderID for domain
		Amount:    getInt64(responseMap, "amount"),
		Status:    mapRefundStatus(getString(responseMap, "status")),
		CreatedAt: getTime(responseMap, "created_at"),
	}

	return refund, nil
}

// ListRefunds retrieves all refunds for an order.
// It takes a GetOrderRequest containing the order ID and returns a slice of canonical Refund domain objects.
func (a *Adapter) ListRefunds(ctx context.Context, req *domain.GetOrderRequest) ([]*domain.Refund, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Build parameters to filter refunds
	params := map[string]interface{}{
		"payment_id": req.OrderID,
	}

	// Call Razorpay SDK to fetch refunds
	// Razorpay All refund method signature: All(params map[string]interface{}, options map[string]string)
	refundsData, err := a.client.Refund.All(params, nil)
	if err != nil {
		// Check if payment not found
		if err.Error() == "Payment not found" {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}

	// Handle the response - Razorpay returns a map with "items" key containing refund list
	itemsList, ok := refundsData["items"].([]interface{})
	if !ok {
		// No items found, return empty slice
		return []*domain.Refund{}, nil
	}

	// Map each refund response to canonical domain type
	var refunds []*domain.Refund
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		refund := &domain.Refund{
			ID:        getString(itemMap, "id"),
			OrderID:   getString(itemMap, "payment_id"), // Razorpay returns payment_id, map to OrderID for domain
			Amount:    getInt64(itemMap, "amount"),
			Status:    mapRefundStatus(getString(itemMap, "status")),
			CreatedAt: getTime(itemMap, "created_at"),
		}

		refunds = append(refunds, refund)
	}

	return refunds, nil
}

// mapRefundStatus converts Razorpay refund status to canonical domain RefundStatus.
// Razorpay uses: "created", "processed", "failed", "partial"
// Domain uses: "created", "processed", "failed", "partial"
func mapRefundStatus(razorpayStatus string) domain.RefundStatus {
	switch razorpayStatus {
	case "created":
		return domain.RefundCreated
	case "processed":
		return domain.RefundProcessed
	case "failed":
		return domain.RefundFailed
	case "partial":
		return domain.RefundPartial
	default:
		return domain.RefundCreated
	}
}
