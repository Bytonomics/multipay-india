package cashfree

import (
	"context"
	"fmt"
	"sync"

	cf "github.com/cashfree/cashfree_pg"

	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// Package-level mutex to guard all Cashfree SDK calls.
// Cashfree SDK uses global variables which are not thread-safe for concurrent clients.
var (
	cfMutex sync.Mutex
)

// Config contains configuration for the Cashfree adapter.
type Config struct {
	// ClientID is the Cashfree merchant client ID.
	ClientID string

	// ClientSecret is the Cashfree merchant client secret.
	ClientSecret string

	// Environment specifies the Cashfree deployment environment (Sandbox or Production).
	Environment domain.Environment
}

// Adapter implements the ProviderAdapter interface for Cashfree payments.
type Adapter struct {
	config *Config
}

// Compile-time assertion that Adapter implements ProviderAdapter interface.
var _ ports.ProviderAdapter = (*Adapter)(nil)

// NewAdapter creates a new Cashfree adapter with the given configuration.
// Returns an error if the configuration is invalid.
func NewAdapter(config *Config) (*Adapter, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required: %w", domain.ErrInvalidRequest)
	}

	if config.ClientID == "" {
		return nil, fmt.Errorf("ClientID is required: %w", domain.ErrInvalidRequest)
	}

	if config.ClientSecret == "" {
		return nil, fmt.Errorf("ClientSecret is required: %w", domain.ErrInvalidRequest)
	}

	if !config.Environment.IsValid() {
		return nil, fmt.Errorf("Environment is invalid: %w", domain.ErrInvalidRequest)
	}

	return &Adapter{
		config: config,
	}, nil
}

// ProviderName returns the provider identifier for Cashfree.
func (a *Adapter) ProviderName() domain.Provider {
	return domain.ProviderCashfree
}

// ProviderCapabilities returns the list of capabilities supported by Cashfree.
// Cashfree supports 8 core capabilities verified across both providers,
// plus additional Cashfree-specific capabilities.
func (a *Adapter) ProviderCapabilities() []domain.Capability {
	return []domain.Capability{
		// Core verified shared capabilities (supported by both Cashfree and Razorpay)
		domain.CapOrderCreate,
		domain.CapOrderFetch,
		domain.CapPaymentFetch,
		domain.CapPaymentList,
		domain.CapRefundCreate,
		domain.CapRefundFetch,
		domain.CapRefundList,
		domain.CapInstrumentFetch,

		// Cashfree-specific capabilities (not in core verified overlap)
		domain.CapInstrumentCryptogram,
		domain.CapOfferCreate,
		domain.CapOfferFetch,
		domain.CapEligibilityFetch,
	}
}

// lockCashfreeSDK acquires the package-level mutex and sets up Cashfree SDK globals.
// This must be paired with unlockCashfreeSDK() in a defer statement.
func (a *Adapter) lockCashfreeSDK() {
	cfMutex.Lock()

	// Set global SDK configuration variables
	clientID := a.config.ClientID
	clientSecret := a.config.ClientSecret

	cf.XClientId = &clientID
	cf.XClientSecret = &clientSecret
	cf.XEnvironment = cf.SANDBOX
	if a.config.Environment == domain.EnvironmentProduction {
		cf.XEnvironment = cf.PRODUCTION
	}

	// Disable Sentry to suppress error reporting side effects
	cf.XEnableErrorAnalytics = false
}

// unlockCashfreeSDK releases the package-level mutex and restores SDK globals to safe defaults.
func (a *Adapter) unlockCashfreeSDK() {
	defer cfMutex.Unlock()

	// Restore globals to safe defaults
	cf.XClientId = nil
	cf.XClientSecret = nil
	cf.XEnvironment = cf.SANDBOX
	cf.XEnableErrorAnalytics = false
}

// CreateOrder creates a new order on the Cashfree payment gateway.
// See orders.go for implementation.
func (a *Adapter) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	return createOrder(ctx, a, req)
}

// GetOrder retrieves an existing order from the Cashfree payment gateway.
// See orders.go for implementation.
func (a *Adapter) GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error) {
	return getOrder(ctx, a, req)
}

// ListOrderPayments retrieves all payments for a specific order.
// See orders.go for implementation.
func (a *Adapter) ListOrderPayments(ctx context.Context, req *domain.GetOrderRequest) ([]*domain.Payment, error) {
	return listOrderPayments(ctx, a, req)
}

// GetPayment retrieves a specific payment for an order.
// See payments.go for implementation.
func (a *Adapter) GetPayment(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	return getPayment(ctx, a, req)
}

// ListPayments retrieves all payments for an order.
// See payments.go for implementation.
func (a *Adapter) ListPayments(ctx context.Context, req *domain.GetOrderRequest) ([]*domain.Payment, error) {
	return listPayments(ctx, a, req)
}

// CreateRefund creates a new refund for an order.
// See refunds.go for implementation.
func (a *Adapter) CreateRefund(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	return createRefund(ctx, a, req)
}

// GetRefund retrieves a specific refund.
// See refunds.go for implementation.
func (a *Adapter) GetRefund(ctx context.Context, req *domain.GetRefundRequest) (*domain.Refund, error) {
	return getRefund(ctx, a, req)
}

// ListRefunds retrieves all refunds for an order.
// See refunds.go for implementation.
func (a *Adapter) ListRefunds(ctx context.Context, req *domain.GetOrderRequest) ([]*domain.Refund, error) {
	return listRefunds(ctx, a, req)
}

// GetInstrument retrieves a specific payment instrument.
// See instruments.go for implementation.
func (a *Adapter) GetInstrument(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	return getInstrument(ctx, a, req)
}

// ListInstruments retrieves all instruments for a customer.
// See instruments.go for implementation.
func (a *Adapter) ListInstruments(ctx context.Context, req *domain.GetInstrumentRequest) ([]*domain.Instrument, error) {
	return listInstruments(ctx, a, req)
}

// DeleteInstrument removes a payment instrument.
// See instruments.go for implementation.
func (a *Adapter) DeleteInstrument(ctx context.Context, req *domain.GetInstrumentRequest) error {
	return deleteInstrument(ctx, a, req)
}

// CreatePaymentLink creates a new shareable payment link.
// See payment_links.go for implementation.
func (a *Adapter) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	return createPaymentLink(ctx, a, req)
}

// GetPaymentLink retrieves an existing payment link.
// See payment_links.go for implementation.
func (a *Adapter) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	return getPaymentLink(ctx, a, req)
}

// CancelPaymentLink cancels an existing payment link.
// See payment_links.go for implementation.
func (a *Adapter) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	return cancelPaymentLink(ctx, a, req)
}

// VerifySignature verifies the authenticity of a webhook request.
// See webhooks.go for implementation.
func (a *Adapter) VerifySignature(ctx context.Context, signature string, payload []byte) error {
	return verifySignature(payload, map[string]string{"X-Cashfree-Signature": signature}, a.config.ClientSecret)
}

// ParseEvent parses and unmarshals a webhook payload into a domain event.
// See webhooks.go for implementation.
func (a *Adapter) ParseEvent(ctx context.Context, payload []byte) (*domain.WebhookEvent, error) {
	return parseEvent(ctx, payload, nil)
}
