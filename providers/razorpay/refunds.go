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
	if req.AmountMinor > 0 {
		params["amount"] = int64(req.AmountMinor)
	}

	// Add notes if provided
	if req.Reason != "" {
		params["notes"] = map[string]string{
			"note": req.Reason,
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
		ProviderRefundID: getString(responseMap, "id"),
		PaymentID:        getString(responseMap, "payment_id"),
		OrderID:          getString(responseMap, "payment_id"), // Razorpay returns payment_id, map to OrderID for domain
		AmountMinor:      domain.AmountMinor(getInt64(responseMap, "amount")),
		Currency:         domain.Currency(getString(responseMap, "currency")),
		Reason:           getString(responseMap, "notes"),
		ARN:              getString(responseMap, "arn"),
		Status:           mapRefundStatus(getString(responseMap, "status")),
		CreatedAt:        getTime(responseMap, "created_at"),
		ProcessedAt:      getTime(responseMap, "receipt_time"),
		Raw:              rawMapResponse(responseMap),
		ProviderDetails: &domain.RefundProviderDetail{
			Razorpay: &domain.RazorpayRefundDetail{
				Entity:         getString(responseMap, "entity"),
				Receipt:        getString(responseMap, "receipt"),
				SpeedRequested: getString(responseMap, "speed_requested"),
				SpeedProcessed: getString(responseMap, "speed_processed"),
				BatchID:        getString(responseMap, "batch_id"),
			},
		},
	}

	if acqData := getMap(responseMap, "acquirer_data"); len(acqData) > 0 {
		refund.ProviderDetails.Razorpay.AcquirerData = &domain.RazorpayAcquirerData{
			BankTransactionID: getString(acqData, "bank_transaction_id"),
			AuthCode:          getString(acqData, "auth_code"),
			RRN:               getString(acqData, "rrn"),
		}
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
		if isNotFoundError(err) {
			return nil, fmt.Errorf("refund %s not found: %w", req.RefundID, domain.ErrRefundNotFound)
		}
		return nil, fmt.Errorf("failed to fetch refund: %w", err)
	}

	// Map Razorpay response to canonical domain type
	refund := &domain.Refund{
		ProviderRefundID: getString(responseMap, "id"),
		PaymentID:        getString(responseMap, "payment_id"),
		OrderID:          getString(responseMap, "payment_id"), // Razorpay returns payment_id, map to OrderID for domain
		AmountMinor:      domain.AmountMinor(getInt64(responseMap, "amount")),
		Currency:         domain.Currency(getString(responseMap, "currency")),
		Reason:           getString(responseMap, "notes"),
		ARN:              getString(responseMap, "arn"),
		Status:           mapRefundStatus(getString(responseMap, "status")),
		CreatedAt:        getTime(responseMap, "created_at"),
		ProcessedAt:      getTime(responseMap, "receipt_time"),
		Raw:              rawMapResponse(responseMap),
		ProviderDetails: &domain.RefundProviderDetail{
			Razorpay: &domain.RazorpayRefundDetail{
				Entity:         getString(responseMap, "entity"),
				Receipt:        getString(responseMap, "receipt"),
				SpeedRequested: getString(responseMap, "speed_requested"),
				SpeedProcessed: getString(responseMap, "speed_processed"),
				BatchID:        getString(responseMap, "batch_id"),
			},
		},
	}

	if acqData := getMap(responseMap, "acquirer_data"); len(acqData) > 0 {
		refund.ProviderDetails.Razorpay.AcquirerData = &domain.RazorpayAcquirerData{
			BankTransactionID: getString(acqData, "bank_transaction_id"),
			AuthCode:          getString(acqData, "auth_code"),
			RRN:               getString(acqData, "rrn"),
		}
	}

	return refund, nil
}

// ListRefunds retrieves all refunds for an order.
// It takes a ListRefundsRequest containing the order ID and returns a slice of canonical Refund domain objects.
func (a *Adapter) ListRefunds(ctx context.Context, req *domain.ListRefundsRequest) ([]*domain.Refund, error) {
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
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment %s not found: %w", req.OrderID, domain.ErrPaymentNotFound)
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
	refunds := make([]*domain.Refund, 0, len(itemsList))
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		refund := &domain.Refund{
			ProviderRefundID: getString(itemMap, "id"),
			PaymentID:        getString(itemMap, "payment_id"),
			OrderID:          getString(itemMap, "payment_id"), // Razorpay returns payment_id, map to OrderID for domain
			AmountMinor:      domain.AmountMinor(getInt64(itemMap, "amount")),
			Currency:         domain.Currency(getString(itemMap, "currency")),
			Reason:           getString(itemMap, "notes"),
			ARN:              getString(itemMap, "arn"),
			Status:           mapRefundStatus(getString(itemMap, "status")),
			CreatedAt:        getTime(itemMap, "created_at"),
			ProcessedAt:      getTime(itemMap, "receipt_time"),
			Raw:              rawMapResponse(itemMap),
			ProviderDetails: &domain.RefundProviderDetail{
				Razorpay: &domain.RazorpayRefundDetail{
					Entity:         getString(itemMap, "entity"),
					Receipt:        getString(itemMap, "receipt"),
					SpeedRequested: getString(itemMap, "speed_requested"),
					SpeedProcessed: getString(itemMap, "speed_processed"),
					BatchID:        getString(itemMap, "batch_id"),
				},
			},
		}

		if acqData := getMap(itemMap, "acquirer_data"); len(acqData) > 0 {
			refund.ProviderDetails.Razorpay.AcquirerData = &domain.RazorpayAcquirerData{
				BankTransactionID: getString(acqData, "bank_transaction_id"),
				AuthCode:          getString(acqData, "auth_code"),
				RRN:               getString(acqData, "rrn"),
			}
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
