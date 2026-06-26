package razorpay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// D17: Typed request structs for Razorpay Refund APIs.
type razorpayCreateRefundRequest struct {
	PaymentID string            `json:"payment_id"`
	Amount    int64             `json:"amount,omitempty"`
	Receipt   string            `json:"receipt,omitempty"`
	Notes     map[string]string `json:"notes,omitempty"`
}

// D17: Typed response struct for Razorpay Refund API responses.
type razorpayRefundResponse struct {
	ID             string                `json:"id"`
	Entity         string                `json:"entity"`
	PaymentID      string                `json:"payment_id"`
	Amount         int64                 `json:"amount"`
	Currency       string                `json:"currency"`
	Status         string                `json:"status"`
	CreatedAt      int64                 `json:"created_at"`
	ReceiptTime    int64                 `json:"receipt_time"`
	Notes          string                `json:"notes"`
	ARN            string                `json:"arn"`
	Receipt        string                `json:"receipt"`
	SpeedRequested string                `json:"speed_requested"`
	SpeedProcessed string                `json:"speed_processed"`
	BatchID        string                `json:"batch_id"`
	AcquirerData   *razorpayAcquirerData `json:"acquirer_data"`
}

// D17: Typed response struct for refund list response.
type razorpayRefundListResponse struct {
	Items []razorpayRefundResponse `json:"items"`
}

// CreateRefund creates a new refund for a payment.
// It takes a CreateRefundRequest with payment ID and optional amount, and returns a canonical Refund domain object.
func (a *Adapter) CreateRefund(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.PaymentID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Build Razorpay refund creation parameters
	refundReq := &razorpayCreateRefundRequest{
		PaymentID: req.PaymentID,
		Amount:    int64(req.AmountMinor), // 0 = full refund; omitempty drops it
		Receipt:   req.RefundID,
	}
	// Merge metadata with reason as "note" key
	if len(req.Metadata) > 0 || req.Reason != "" {
		notes := make(map[string]string)
		for k, v := range req.Metadata {
			notes[k] = v
		}
		if req.Reason != "" {
			notes["note"] = req.Reason
		}
		refundReq.Notes = notes
	}
	params, err := encodeRequest(refundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode create refund request: %w", err)
	}

	// Call Razorpay SDK to create refund
	// Razorpay Create refund method signature: Create(params map[string]interface{}, options map[string]string)
	responseMap, err := a.client.Refund.Create(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayRefundResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapRefundFromResponse(typed, rawJSON), nil
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

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayRefundResponse](responseMap)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed struct to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund response: %w", err)
	}

	// D17: Map typed struct to canonical domain type
	return mapRefundFromResponse(typed, rawJSON), nil
}

// ListRefunds retrieves every refund issued against a Razorpay PAYMENT. Razorpay scopes
// refunds under the captured payment (not the order), so this uses the canonical
// ListRefundsRequest.PaymentID and returns a slice of canonical Refund domain objects.
func (a *Adapter) ListRefunds(ctx context.Context, req *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	if req == nil {
		return nil, domain.ErrInvalidRequest
	}
	if req.PaymentID == "" {
		return nil, domain.ErrInvalidRequest
	}

	// Razorpay lists refunds per payment: GET /v1/payments/{payment_id}/refunds.
	// NOTE: the account-wide Refund.All endpoint does NOT filter by payment id, so we must
	// use the payment-scoped FetchMultipleRefund here.
	refundsData, err := a.client.Payment.FetchMultipleRefund(req.PaymentID, nil, nil)
	if err != nil {
		// Check if payment not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment %s not found: %w", req.PaymentID, domain.ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}

	// D17: Decode map to typed struct at SDK boundary
	typed, err := decodeResponse[razorpayRefundListResponse](refundsData)
	if err != nil {
		return nil, err
	}

	// D17: Marshal typed list to bytes for mapper
	rawJSON, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refunds response: %w", err)
	}

	// D17: Map each typed refund response to canonical domain type
	refunds := make([]*domain.Refund, 0, len(typed.Items))
	for i := range typed.Items {
		refund := mapRefundFromResponse(&typed.Items[i], rawJSON)
		refunds = append(refunds, refund)
	}

	return refunds, nil
}

// D17: Typed struct mapper for refund response
func mapRefundFromResponse(r *razorpayRefundResponse, rawJSON []byte) *domain.Refund {
	refund := &domain.Refund{
		ProviderRefundID: r.ID,
		PaymentID:        r.PaymentID,
		OrderID:          r.PaymentID, // Razorpay returns payment_id, map to OrderID for domain
		AmountMinor:      domain.AmountMinor(r.Amount),
		Currency:         domain.Currency(r.Currency),
		Reason:           r.Notes,
		ARN:              r.ARN,
		Status:           mapRefundStatus(r.Status),
		CreatedAt:        unixPtr(r.CreatedAt),
		ProcessedAt:      unixPtr(r.ReceiptTime),
		Raw:              domain.RawProviderResponse(rawJSON),
		ProviderDetails: &domain.RefundProviderDetail{
			Razorpay: &domain.RazorpayRefundDetail{
				Entity:         r.Entity,
				Receipt:        r.Receipt,
				SpeedRequested: r.SpeedRequested,
				SpeedProcessed: r.SpeedProcessed,
				BatchID:        r.BatchID,
			},
		},
	}

	// D17: Map acquirer_data if present
	if r.AcquirerData != nil {
		refund.ProviderDetails.Razorpay.AcquirerData = &domain.RazorpayAcquirerData{
			BankTransactionID: r.AcquirerData.BankTransactionID,
			AuthCode:          r.AcquirerData.AuthCode,
			RRN:               r.AcquirerData.RRN,
		}
	}

	return refund
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
