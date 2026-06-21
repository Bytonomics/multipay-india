package client

import (
	"errors"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/orchestration"
	"github.com/Bytonomics/multipay-adapter/ports"
	"github.com/Bytonomics/multipay-adapter/routing"
)

// MultiPayClient is the main public API for the payment adapter.
// It provides orchestration services for orders, payments, refunds, instruments,
// payment links, webhooks, and capability queries across multiple payment providers.
//
// All services share a single ProviderRegistry, SupportMatrix, and HookPipeline
// for consistent behavior across the client.
type MultiPayClient struct {
	orders       *orchestration.OrderService
	payments     *orchestration.PaymentService
	refunds      *orchestration.RefundService
	instruments  *orchestration.InstrumentService
	paymentLinks *orchestration.PaymentLinkService
	webhooks     *orchestration.WebhookService
	capabilities *orchestration.CapabilityService
}

// NewClient creates a new MultiPayClient from the provided configuration.
//
// The client wires up all dependencies:
// - Validates the config
// - Creates a capability support matrix from provider declarations
// - Creates a capability validator for early unsupported-capability detection
// - Creates and populates a provider registry with all configured adapters
// - Creates a hook pipeline from configured hooks
// - Creates all 7 orchestration services with shared dependencies
// - Optionally creates a webhook endpoint registry for webhook routing
//
// Returns an error if the config is invalid or if provider registration fails.
func NewClient(cfg *ClientConfig) (*MultiPayClient, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create capability support matrix (pre-populated with both provider capabilities)
	supportMatrix := capabilities.NewSupportMatrix()

	// Create capability validator for early validation before adapter dispatch
	validator := capabilities.NewValidator(supportMatrix)

	// Create provider registry and register all configured adapters
	registry := ports.NewProviderRegistry()
	for i, adapter := range cfg.Providers {
		providerName := adapter.ProviderName()
		if err := registry.Register(providerName, adapter); err != nil {
			return nil, fmt.Errorf("failed to register provider at index %d: %w", i, err)
		}
	}

	// Create hook pipeline from configured hooks (empty if none provided)
	var hookList []ports.Hook
	if cfg.Hooks != nil {
		hookList = cfg.Hooks
	}
	pipeline := hooks.NewPipeline(hookList...)

	// Create all 7 orchestration services with shared registry, validator, and pipeline
	orderService := orchestration.NewOrderService(registry, validator, pipeline)
	paymentService := orchestration.NewPaymentService(registry, validator, pipeline)
	refundService := orchestration.NewRefundService(registry, validator, pipeline)
	instrumentService := orchestration.NewInstrumentService(registry, validator, pipeline)
	paymentLinkService := orchestration.NewPaymentLinkService(registry, validator, pipeline)

	// WebhookService has a different constructor (requires EndpointRegistry, Store, Logger)
	endpointRegistry := routing.NewEndpointRegistry()
	webhookService := orchestration.NewWebhookService(endpointRegistry, cfg.WebhookStore, cfg.Logger)

	// CapabilityService delegates to SupportMatrix
	capabilityService := orchestration.NewCapabilityService(supportMatrix)

	return &MultiPayClient{
		orders:       orderService,
		payments:     paymentService,
		refunds:      refundService,
		instruments:  instrumentService,
		paymentLinks: paymentLinkService,
		webhooks:     webhookService,
		capabilities: capabilityService,
	}, nil
}

// Orders returns the OrderService for order operations.
func (c *MultiPayClient) Orders() *orchestration.OrderService {
	return c.orders
}

// Payments returns the PaymentService for payment operations.
func (c *MultiPayClient) Payments() *orchestration.PaymentService {
	return c.payments
}

// Refunds returns the RefundService for refund operations.
func (c *MultiPayClient) Refunds() *orchestration.RefundService {
	return c.refunds
}

// Instruments returns the InstrumentService for payment instrument operations.
func (c *MultiPayClient) Instruments() *orchestration.InstrumentService {
	return c.instruments
}

// PaymentLinks returns the PaymentLinkService for payment link operations.
func (c *MultiPayClient) PaymentLinks() *orchestration.PaymentLinkService {
	return c.paymentLinks
}

// Webhooks returns the WebhookService for webhook event processing.
func (c *MultiPayClient) Webhooks() *orchestration.WebhookService {
	return c.webhooks
}

// Capabilities returns the CapabilityService for capability discovery and queries.
func (c *MultiPayClient) Capabilities() *orchestration.CapabilityService {
	return c.capabilities
}
