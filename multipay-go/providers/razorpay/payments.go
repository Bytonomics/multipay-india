package razorpay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// D17: Typed request struct for Razorpay List Payments API.
type razorpayListPaymentsRequest struct {
	OrderID string `json:"order_id"`
}

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

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapPaymentFromResponse(typed, rawJSON), nil
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
	params, err := encodeRequest(&razorpayListPaymentsRequest{OrderID: req.OrderID})
	if err != nil {
		return nil, fmt.Errorf("failed to encode list payments request: %w", err)
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

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentListResponse](paymentsData)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed list to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payments response: %w", err)
	}

	// D17: Map each typed payment response to canonical domain type
	payments := make([]*domain.Payment, 0, len(typed.Items))
	for i := range typed.Items {
		payment := mapPaymentFromResponse(&typed.Items[i], rawJSON)
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

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayPaymentResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapPaymentFromResponse(typed, rawJSON), nil
}
