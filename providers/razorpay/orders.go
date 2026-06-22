package razorpay

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// CreateOrder creates a new order on Razorpay.
// It takes a CreateOrderRequest and returns a canonical Order domain object.
// The amount is in paisa (minor currency unit), which Razorpay uses natively.
func (a *Adapter) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}

	// Build Razorpay order creation parameters
	params := map[string]interface{}{
		"amount":   req.AmountMinor, // paisa
		"currency": string(req.Currency),
		"notes":    req.Metadata,
	}

	// Add optional order ID as receipt if provided
	if req.OrderID != "" {
		params["receipt"] = req.OrderID
	}

	// Call Razorpay SDK to create order
	responseMap, err := a.client.Order.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Map Razorpay response to canonical domain type
	order := &domain.Order{
		ProviderOrderID: getString(responseMap, "id"),
		OrderID:         getString(responseMap, "receipt"),
		AmountMinor:     domain.AmountMinor(getInt64(responseMap, "amount")),
		Currency:        domain.Currency(getString(responseMap, "currency")),
		Status:          mapOrderStatus(getString(responseMap, "status")),
		CreatedAt:       getTime(responseMap, "created_at"),
		Raw:             rawMapResponse(responseMap),
	}

	return order, nil
}

// GetOrder retrieves an existing order from Razorpay.
// It takes a GetOrderRequest and returns a canonical Order domain object.
func (a *Adapter) GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch order
	responseMap, err := a.client.Order.Fetch(req.OrderID, nil, nil)
	if err != nil {
		// Check if order not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("order %s not found: %w", req.OrderID, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	// Map Razorpay response to canonical domain type
	order := &domain.Order{
		ProviderOrderID: getString(responseMap, "id"),
		OrderID:         getString(responseMap, "receipt"),
		AmountMinor:     domain.AmountMinor(getInt64(responseMap, "amount")),
		Currency:        domain.Currency(getString(responseMap, "currency")),
		Status:          mapOrderStatus(getString(responseMap, "status")),
		CreatedAt:       getTime(responseMap, "created_at"),
		Raw:             rawMapResponse(responseMap),
	}

	return order, nil
}

// ListOrderPayments retrieves all payments associated with a specific order.
// It takes a ListOrderPaymentsRequest with order ID and returns a slice of canonical Payment domain objects.
func (a *Adapter) ListOrderPayments(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Call Razorpay SDK to fetch payments for the order
	paymentsData, err := a.client.Order.Payments(req.OrderID, nil, nil)
	if err != nil {
		// Check if order not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("order %s not found: %w", req.OrderID, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("failed to list order payments: %w", err)
	}

	// Handle the response - Razorpay returns a map with "items" key containing payment list
	items := getMap(paymentsData, "items")

	// Check if items is actually a list (slice)
	itemsList, ok := items["items"].([]interface{})
	if !ok {
		// Try to extract items directly if it's the top-level response
		if topItems, ok := paymentsData["items"].([]interface{}); ok {
			itemsList = topItems
		} else {
			// No items found, return empty slice
			return []*domain.Payment{}, nil
		}
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
			Status:            mapPaymentStatus(getString(itemMap, "status")),
			PaymentMethod:     getString(itemMap, "method"),
			PaymentTime:       getTime(itemMap, "created_at"),
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// mapOrderStatus converts Razorpay order status to canonical domain OrderStatus.
// Razorpay uses: "created", "attempted", "paid", "cancelled", "expired"
// Domain uses: "created", "paid", "expired", "cancelled"
func mapOrderStatus(razorpayStatus string) domain.OrderStatus {
	switch razorpayStatus {
	case "created":
		return domain.OrderCreated
	case "attempted":
		// Razorpay has "attempted" state when payment is being processed
		// For canonical domain, we treat this as "created" (not yet paid)
		return domain.OrderCreated
	case "paid":
		return domain.OrderPaid
	case "expired":
		return domain.OrderExpired
	case "cancelled":
		return domain.OrderCancelled
	default:
		return domain.OrderCreated
	}
}

// mapPaymentStatus converts Razorpay payment status to canonical domain PaymentStatus.
// Razorpay uses: "authorized", "captured", "failed", "refunded"
// Domain uses: "authorized", "captured", "failed", "refunded"
func mapPaymentStatus(razorpayStatus string) domain.PaymentStatus {
	switch razorpayStatus {
	case "authorized":
		return domain.PaymentAuthorized
	case "captured":
		return domain.PaymentCaptured
	case "failed":
		return domain.PaymentFailed
	case "refunded":
		return domain.PaymentRefunded
	default:
		return domain.PaymentFailed
	}
}
