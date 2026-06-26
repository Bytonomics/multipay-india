package cashfree

import (
	"context"
	"fmt"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// createPaymentLink creates a new shareable payment link on the Cashfree payment gateway.
// Maps the canonical domain.CreatePaymentLinkRequest to a Cashfree CreateLinkRequest,
// calls the SDK, and maps the response back to a canonical domain.PaymentLink.
func createPaymentLink(ctx context.Context, adapter *Adapter, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.AmountMinor <= 0 {
		return nil, fmt.Errorf("amount must be positive: %w", domain.ErrInvalidRequest)
	}

	if req.Currency == "" {
		return nil, fmt.Errorf("currency is required: %w", domain.ErrInvalidRequest)
	}

	// Build Cashfree CreateLinkRequest
	cfReq := &cf.CreateLinkRequest{
		LinkAmount:   AmountMinorToMajor(int64(req.AmountMinor), string(req.Currency)),
		LinkCurrency: string(req.Currency),
		LinkPurpose:  stringPtr(req.Purpose),
	}

	// Call Cashfree SDK
	cfLink, httpResp, err := adapter.cfClient.PGCreateLinkWithContext(
		ctx,
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
		return nil, fmt.Errorf("failed to create payment link on cashfree: %w", domain.ErrProviderError)
	}

	if cfLink == nil {
		return nil, fmt.Errorf("cashfree returned nil payment link: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	link, err := MapLinkEntityToCanonical(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to map link: %w", err)
	}
	return link, nil
}

// getPaymentLink retrieves an existing payment link from the Cashfree payment gateway.
// Maps the canonical domain.GetPaymentLinkRequest to a Cashfree fetch request,
// calls the SDK, and maps the response back to a canonical domain.PaymentLink.
func getPaymentLink(ctx context.Context, adapter *Adapter, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.LinkID == "" {
		return nil, fmt.Errorf("LinkID is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK to fetch payment link
	cfLink, httpResp, err := adapter.cfClient.PGFetchLinkWithContext(
		ctx,
		req.LinkID,
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
		// Check if error is 404 link not found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("payment link %s not found: %w", req.LinkID, domain.ErrPaymentLinkNotFound)
		}
		return nil, fmt.Errorf("failed to fetch payment link from Cashfree: %w", domain.ErrProviderError)
	}

	if cfLink == nil {
		return nil, fmt.Errorf("cashfree returned nil payment link: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	link, err := MapLinkEntityToCanonical(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to map link: %w", err)
	}
	return link, nil
}

// cancelPaymentLink cancels an existing payment link on the Cashfree payment gateway.
// Maps the canonical domain.CancelPaymentLinkRequest to a Cashfree cancel request,
// calls the SDK, and maps the response back to a canonical domain.PaymentLink.
func cancelPaymentLink(ctx context.Context, adapter *Adapter, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required: %w", domain.ErrInvalidRequest)
	}

	if req.LinkID == "" {
		return nil, fmt.Errorf("LinkID is required: %w", domain.ErrInvalidRequest)
	}

	// Call Cashfree SDK to cancel payment link
	cfLink, httpResp, err := adapter.cfClient.PGCancelLinkWithContext(
		ctx,
		req.LinkID,
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
		return nil, fmt.Errorf("failed to cancel payment link on cashfree: %w", domain.ErrProviderError)
	}

	if cfLink == nil {
		return nil, fmt.Errorf("cashfree returned nil payment link: %w", domain.ErrProviderError)
	}

	// Map response to canonical type
	link, err := MapLinkEntityToCanonical(cfLink)
	if err != nil {
		return nil, fmt.Errorf("failed to map link: %w", err)
	}
	return link, nil
}
