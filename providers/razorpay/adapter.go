package razorpay

import (
	"context"
	"errors"

	"github.com/razorpay/razorpay-go"

	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// Config holds the Razorpay API credentials.
type Config struct {
	// Key is the Razorpay public key (API key).
	Key string

	// Secret is the Razorpay secret key.
	Secret string
}

// Adapter implements the ProviderAdapter interface for Razorpay.
// It wraps the Razorpay SDK client and provides all payment operations.
type Adapter struct {
	client *razorpay.Client
	config *Config
}

// Compile-time check to ensure Adapter implements ProviderAdapter.
var _ ports.ProviderAdapter = (*Adapter)(nil)

// NewAdapter creates a new Razorpay adapter with the given configuration.
// It initializes the Razorpay SDK client with the provided credentials.
// Returns an error if the configuration is invalid.
func NewAdapter(cfg *Config) (*Adapter, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if cfg.Key == "" {
		return nil, errors.New("razorpay key cannot be empty")
	}
	if cfg.Secret == "" {
		return nil, errors.New("razorpay secret cannot be empty")
	}

	// Create Razorpay client with key and secret
	rzClient := razorpay.NewClient(cfg.Key, cfg.Secret)

	adapter := &Adapter{
		client: rzClient,
		config: cfg,
	}

	return adapter, nil
}

// ProviderName returns the provider identifier for Razorpay.
func (a *Adapter) ProviderName() domain.Provider {
	return domain.ProviderRazorpay
}

// ProviderCapabilities returns all capabilities supported by Razorpay.
// This includes both core shared capabilities and Razorpay-specific capabilities.
// Returns all 27 Razorpay-supported capabilities.
func (a *Adapter) ProviderCapabilities() []domain.Capability {
	return []domain.Capability{
		// Core shared capabilities (14)
		domain.CapOrderCreate,
		domain.CapOrderFetch,
		domain.CapPaymentFetch,
		domain.CapPaymentList,
		domain.CapPaymentPay,
		domain.CapRefundCreate,
		domain.CapRefundFetch,
		domain.CapRefundList,
		domain.CapInstrumentFetch,
		domain.CapInstrumentList,
		domain.CapInstrumentDelete,
		domain.CapPaymentLinkCreate,
		domain.CapPaymentLinkFetch,
		domain.CapPaymentLinkCancel,

		// Razorpay-specific capabilities (13)
		domain.CapOrderUpdate,
		domain.CapOrderList,
		domain.CapPaymentCapture,
		domain.CapRefundUpdate,
		domain.CapCustomerCreate,
		domain.CapCustomerFetch,
		domain.CapCustomerEdit,
		domain.CapCustomerList,
		domain.CapWebhookCreate,
		domain.CapWebhookFetch,
		domain.CapWebhookEdit,
		domain.CapWebhookDelete,
		domain.CapWebhookList,
		domain.CapSubscriptionCreate,
		domain.CapSubscriptionFetch,
		domain.CapSubscriptionList,
		domain.CapPlanCreate,
		domain.CapPlanFetch,
		domain.CapPlanList,
		domain.CapPaymentLinkUpdate,
		domain.CapPaymentLinkNotify,
		domain.CapPaymentLinkList,
		domain.CapUPICreate,
		domain.CapVPAValidate,
	}
}

// VerifySignature verifies the authenticity of a webhook request from Razorpay.
// See webhooks.go for implementation.
func (a *Adapter) VerifySignature(ctx context.Context, signature string, payload []byte) error {
	return verifySignature(payload, map[string]string{"X-Razorpay-Signature": signature}, a.config.Secret)
}

// ParseEvent parses and unmarshals a Razorpay webhook payload into a domain event.
// See webhooks.go for implementation.
func (a *Adapter) ParseEvent(ctx context.Context, payload []byte) (*domain.WebhookEvent, error) {
	return parseEvent(ctx, payload, nil)
}
