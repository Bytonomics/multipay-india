package ports

import (
	"context"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
)

// OrderProvider defines operations for payment orders.
type OrderProvider interface {
	// CreateOrder creates a new order on the payment provider.
	CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error)

	// GetOrder retrieves an existing order from the payment provider.
	GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error)

	// ListOrderPayments retrieves all payments associated with a specific order.
	ListOrderPayments(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error)
}

// PaymentProvider defines operations for payments.
type PaymentProvider interface {
	// GetPayment retrieves a specific payment for an order.
	GetPayment(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error)

	// ListPayments retrieves all payments for an order.
	ListPayments(ctx context.Context, req *domain.ListPaymentsRequest) ([]*domain.Payment, error)

	// CapturePayment captures an authorized payment.
	CapturePayment(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error)
}

// RefundProvider defines operations for refunds.
type RefundProvider interface {
	// CreateRefund creates a new refund for an order.
	CreateRefund(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error)

	// GetRefund retrieves an existing refund.
	GetRefund(ctx context.Context, req *domain.GetRefundRequest) (*domain.Refund, error)

	// ListRefunds retrieves all refunds for an order.
	ListRefunds(ctx context.Context, req *domain.ListRefundsRequest) ([]*domain.Refund, error)
}

// InstrumentProvider defines operations for payment instruments.
type InstrumentProvider interface {
	// GetInstrument retrieves a specific payment instrument.
	GetInstrument(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error)

	// ListInstruments retrieves all instruments for a customer.
	ListInstruments(ctx context.Context, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error)

	// DeleteInstrument removes a payment instrument.
	DeleteInstrument(ctx context.Context, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error)
}

// PaymentLinkProvider defines operations for payment links.
type PaymentLinkProvider interface {
	// CreatePaymentLink creates a new shareable payment link.
	CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error)

	// GetPaymentLink retrieves an existing payment link.
	GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error)

	// CancelPaymentLink cancels an existing payment link.
	CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error)
}

// WebhookConsumerProvider defines operations for webhook processing.
type WebhookConsumerProvider interface {
	// VerifySignature verifies the authenticity of a webhook request.
	// Returns an error if verification fails.
	VerifySignature(ctx context.Context, payload []byte, headers map[string]string) error

	// ParseEvent parses and unmarshals a webhook payload into a domain event.
	ParseEvent(ctx context.Context, payload []byte, headers map[string]string) (*domain.WebhookEvent, error)
}

// ProviderAdapter is the main port that embeds all provider-specific capability interfaces.
// An adapter implementing this interface provides all payment operations for a specific provider.
// It also includes methods for provider identification and capability discovery.
type ProviderAdapter interface {
	OrderProvider
	PaymentProvider
	RefundProvider
	InstrumentProvider
	PaymentLinkProvider
	WebhookConsumerProvider
	MetadataMapper

	// ProviderName returns the provider identifier for this adapter.
	ProviderName() domain.Provider

	// ProviderCapabilities returns the list of capabilities supported by this provider.
	ProviderCapabilities() []capabilities.Capability
}
