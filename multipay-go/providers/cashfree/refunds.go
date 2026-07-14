package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
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

	// Build Cashfree refund request
	refundAmount := 0.0
	if req.AmountMinor > 0 {
		refundAmount = currencyutils.AmountMinorToMajor(int64(req.AmountMinor), string(req.Currency))
	}

	cfReq := &cf.OrderCreateRefundRequest{
		RefundAmount: refundAmount,
	}

	// Guard optional RefundNote (only if Reason is non-empty)
	if req.Reason != "" {
		reason := req.Reason
		cfReq.RefundNote = &reason
	}

	if req.RefundID != "" {
		refundId := req.RefundID
		cfReq.RefundId = &refundId
	}

	// Forward optional refund_speed (STANDARD | INSTANT). Canonical values map 1:1 to Cashfree's.
	if req.RefundSpeed != "" {
		speed := string(req.RefundSpeed)
		cfReq.RefundSpeed = &speed
	}

	// Forward optional refund_splits (reverse an Easy-Split order across vendors).
	if len(req.RefundSplits) > 0 {
		splits := make([]cf.OrderCreateRefundRequestRefundSplitsInner, 0, len(req.RefundSplits))
		for i := range req.RefundSplits {
			s := &req.RefundSplits[i]
			inner := cf.OrderCreateRefundRequestRefundSplitsInner{
				VendorId: s.VendorID,
			}
			if s.Amount != 0 {
				amt := float32(currencyutils.AmountMinorToMajor(int64(s.Amount), string(req.Currency)))
				inner.Amount = &amt
			}
			if len(s.Tags) > 0 {
				inner.Tags = wrapSplitTags(s.Tags)
			}
			splits = append(splits, inner)
		}
		cfReq.RefundSplits = splits
	}

	// Call Cashfree SDK to create refund
	cfRefund, httpResp, err := adapter.cfClient.PGOrderCreateRefundWithContext(
		ctx,
		req.OrderID,
		cfReq,
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
		return nil, fmt.Errorf("failed to create refund on cashfree: %w", domain.ErrProviderError)
	}

	if cfRefund == nil {
		return nil, fmt.Errorf("cashfree returned nil refund: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	refund, err := MapRefundEntityToCanonical(cfRefund)
	if err != nil {
		return nil, fmt.Errorf("failed to map refund: %w", err)
	}
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

	// Call Cashfree SDK to fetch refund
	cfRefund, httpResp, err := adapter.cfClient.PGOrderFetchRefundWithContext(
		ctx,
		req.OrderID,
		req.RefundID,
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
		// Check if error is 404 refund not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("refund %s not found: %w", req.RefundID, domain.ErrRefundNotFound)
		}
		return nil, fmt.Errorf("failed to fetch refund from Cashfree: %w", domain.ErrProviderError)
	}

	if cfRefund == nil {
		return nil, fmt.Errorf("cashfree returned nil refund: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	refund, err := MapRefundEntityToCanonical(cfRefund)
	if err != nil {
		return nil, fmt.Errorf("failed to map refund: %w", err)
	}
	return refund, nil
}

// listRefunds retrieves all refunds for a specific order from the Cashfree payment gateway.
// Calls the Cashfree SDK to fetch refunds and maps them to canonical domain.Refund types.
func listRefunds(ctx context.Context, adapter *Adapter, req *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("OrderID is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK to fetch refunds for the order
	cfRefunds, httpResp, err := adapter.cfClient.PGOrderFetchRefundsWithContext(
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
		return nil, fmt.Errorf("failed to fetch refunds from Cashfree: %w", domain.ErrProviderError)
	}

	if cfRefunds == nil {
		return []*domain.Refund{}, nil
	}

	// Map response to canonical types
	result := make([]*domain.Refund, 0, len(cfRefunds))
	for i := range cfRefunds {
		cfRefund := &cfRefunds[i]
		refund, err := MapRefundEntityToCanonical(cfRefund)
		if err != nil {
			return nil, fmt.Errorf("failed to map refund at index %d: %w", i, err)
		}
		if refund != nil {
			result = append(result, refund)
		}
	}

	return result, nil
}
