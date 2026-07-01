package client

import (
	"errors"
	"fmt"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/orchestration"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/routing"
)

// MultiPayClientInterface defines the contract for payment operations.
// This interface is implemented by MultiPayClient and can be used for mocking in tests.
type MultiPayClientInterface interface {
	Orders() *orchestration.OrderService
	Payments() *orchestration.PaymentService
	Refunds() *orchestration.RefundService
	Instruments() *orchestration.InstrumentService
	PaymentLinks() *orchestration.PaymentLinkService
	Webhooks() *orchestration.WebhookService
	Capabilities() *orchestration.CapabilityService
	Plans() *orchestration.PlanService
	Subscriptions() *orchestration.SubscriptionService
}

// MultiPayClient is the main public API for the payment adapter.
// It provides orchestration services for orders, payments, refunds, instruments,
// payment links, webhooks, and capability queries for a single payment provider.
//
// All services share a single ProviderAdapter, SupportMatrix, and HookPipeline
// for consistent behavior across the client.
type MultiPayClient struct {
	orders        *orchestration.OrderService
	payments      *orchestration.PaymentService
	refunds       *orchestration.RefundService
	instruments   *orchestration.InstrumentService
	paymentLinks  *orchestration.PaymentLinkService
	webhooks      *orchestration.WebhookService
	plans         *orchestration.PlanService
	subscriptions *orchestration.SubscriptionService
	capabilities  *orchestration.CapabilityService
}

// NewClient creates a new MultiPayClient from the provided configuration.
//
// The client wires up all dependencies:
// - Validates the config
// - Derives the configured provider identity from the adapter implementation
// - Creates a capability support matrix from provider declarations
// - Creates a capability validator for early unsupported-capability detection
// - Creates a hook pipeline from configured hooks
// - Creates all 7 orchestration services with shared dependencies
// - Optionally creates a webhook endpoint registry for webhook routing
//
// Returns an error if the config is invalid.
func NewClient(cfg *ClientConfig) (*MultiPayClient, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	if cfg.WebhookStore == nil {
		panic("ClientConfig.WebhookStore is required (cannot be nil); WebhookStore provides durable event capture for webhook replay and idempotency")
	}

	if cfg.Provider == nil {
		panic("ClientConfig.Provider is required (cannot be nil); a payment client must be bound to exactly one provider adapter")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	adapter := cfg.Provider
	provider := adapter.ProviderName()

	// Create capability support matrix (pre-populated with both provider capabilities)
	supportMatrix := capabilities.NewSupportMatrix()

	// Create capability validator for early validation before adapter dispatch
	validator := capabilities.NewValidator(supportMatrix)

	// Use provided logger or noop logger if not configured
	logger := cfg.Logger
	if logger == nil {
		logger = ports.NewNoopLogger()
	}

	// Create hook pipeline from configured hooks (empty if none provided)
	var hookList []ports.Hook
	if cfg.Hooks != nil {
		hookList = cfg.Hooks
	}
	pipeline := hooks.NewPipeline(logger, hookList...)

	// Create clock for time-related operations
	clock := cfg.Clock
	if clock == nil {
		clock = ports.NewRealClock()
	}

	// Create all 7 orchestration services with the configured provider, adapter, validator, and pipeline
	orderService := orchestration.NewOrderService(provider, adapter, validator, pipeline, logger, clock)
	paymentService := orchestration.NewPaymentService(provider, adapter, validator, pipeline, logger, clock)
	refundService := orchestration.NewRefundService(provider, adapter, validator, pipeline, logger, clock)
	instrumentService := orchestration.NewInstrumentService(provider, adapter, validator, pipeline, logger, clock)
	paymentLinkService := orchestration.NewPaymentLinkService(provider, adapter, validator, pipeline, logger, clock)

	// Create PlanService and SubscriptionService with validator for capability checks
	planService := orchestration.NewPlanService(provider, adapter, pipeline, logger, clock)
	subscriptionService := orchestration.NewSubscriptionService(provider, adapter, validator, pipeline, logger, clock)

	// WebhookService has a different constructor (requires Provider, Adapter, Pipeline, Store, EndpointRegistry, Logger)
	endpointRegistry := routing.NewEndpointRegistry()
	webhookService := orchestration.NewWebhookService(provider, adapter, pipeline, cfg.WebhookStore, endpointRegistry, logger)

	// CapabilityService delegates to SupportMatrix
	capabilityService := orchestration.NewCapabilityService(supportMatrix)

	return &MultiPayClient{
		orders:        orderService,
		payments:      paymentService,
		refunds:       refundService,
		instruments:   instrumentService,
		paymentLinks:  paymentLinkService,
		webhooks:      webhookService,
		plans:         planService,
		subscriptions: subscriptionService,
		capabilities:  capabilityService,
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

// Plans returns the PlanService for plan operations.
func (c *MultiPayClient) Plans() *orchestration.PlanService {
	return c.plans
}

// Subscriptions returns the SubscriptionService for subscription operations.
func (c *MultiPayClient) Subscriptions() *orchestration.SubscriptionService {
	return c.subscriptions
}
