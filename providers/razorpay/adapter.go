package razorpay

import (
	"context"
	"errors"

	"github.com/razorpay/razorpay-go"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// Config holds the Razorpay API credentials.
type Config struct {
	// Key is the Razorpay public key (API key).
	Key string

	// Secret is the Razorpay secret key.
	Secret string

	// WebhookSecret is the HMAC-SHA256 secret for webhook verification.
	WebhookSecret string

	// AccountID is the unique account ID for webhook routing.
	AccountID string
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
// Returns all 29 Razorpay-supported capabilities.
func (a *Adapter) ProviderCapabilities() []capabilities.Capability {
	return []capabilities.Capability{
		// Core shared capabilities (16)
		capabilities.CapOrderCreate,
		capabilities.CapOrderFetch,
		capabilities.CapOrderListPayments,
		capabilities.CapPaymentFetch,
		capabilities.CapPaymentList,
		capabilities.CapPaymentPay,
		capabilities.CapRefundCreate,
		capabilities.CapRefundFetch,
		capabilities.CapRefundList,
		capabilities.CapInstrumentFetch,
		capabilities.CapInstrumentList,
		capabilities.CapInstrumentDelete,
		capabilities.CapPaymentLinkCreate,
		capabilities.CapPaymentLinkFetch,
		capabilities.CapPaymentLinkCancel,
		capabilities.CapWebhookConsume,

		// Razorpay-specific capabilities (13)
		capabilities.CapOrderUpdate,
		capabilities.CapOrderList,
		capabilities.CapPaymentCapture,
		capabilities.CapRefundUpdate,
		capabilities.CapCustomerCreate,
		capabilities.CapCustomerFetch,
		capabilities.CapCustomerEdit,
		capabilities.CapCustomerList,
		capabilities.CapWebhookCreate,
		capabilities.CapWebhookFetch,
		capabilities.CapWebhookEdit,
		capabilities.CapWebhookDelete,
		capabilities.CapWebhookList,
		capabilities.CapSubscriptionCreate,
		capabilities.CapSubscriptionFetch,
		capabilities.CapSubscriptionList,
		capabilities.CapPlanCreate,
		capabilities.CapPlanFetch,
		capabilities.CapPlanList,
		capabilities.CapPaymentLinkUpdate,
		capabilities.CapPaymentLinkNotify,
		capabilities.CapPaymentLinkList,
		capabilities.CapUPICreate,
		capabilities.CapVPAValidate,
	}
}

// VerifySignature verifies the authenticity of a webhook request from Razorpay.
// See webhooks.go for implementation.
func (a *Adapter) VerifySignature(ctx context.Context, payload []byte, headers map[string]string) error {
	return verifySignature(payload, headers, a.config.WebhookSecret)
}

// ParseEvent parses and unmarshals a Razorpay webhook payload into a domain event.
// See webhooks.go for implementation.
func (a *Adapter) ParseEvent(ctx context.Context, payload []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	return parseEvent(ctx, payload, headers)
}

func (a *Adapter) MapOrderMetadata(_ context.Context, metadata domain.Metadata) (map[string]interface{}, error) {
	notes := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		notes[k] = v
	}
	return map[string]interface{}{"notes": notes}, nil
}

func (a *Adapter) MapRefundMetadata(_ context.Context, metadata domain.Metadata) (map[string]interface{}, error) {
	notes := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		notes[k] = v
	}
	return map[string]interface{}{"notes": notes}, nil
}

func (a *Adapter) MapPaymentLinkMetadata(_ context.Context, metadata domain.Metadata) (map[string]interface{}, error) {
	notes := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		notes[k] = v
	}
	return map[string]interface{}{"notes": notes}, nil
}
