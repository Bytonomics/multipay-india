package fake

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// compile-time interface check — proves FakeAdapter implements every ProviderAdapter method.
var _ ports.ProviderAdapter = (*FakeAdapter)(nil)

// FakeAdapter is a vendor-free ports.ProviderAdapter. Per-operation behavior is set
// via the *Func fields (nil => a safe zero-value default). Calls are recorded in the
// *Calls slices for assertions. VerifySignature always succeeds and ParseEvent
// decodes the fake's own WebhookEnvelope.
type FakeAdapter struct {
	mu           sync.Mutex
	providerName domain.Provider

	// Configurable behavior (nil => zero-value default).
	CreateOrderFunc             func(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error)
	GetOrderFunc                func(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error)
	ListOrderPaymentsFunc       func(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error)
	GetPaymentFunc              func(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error)
	ListPaymentsFunc            func(ctx context.Context, req *domain.ListPaymentsRequest) ([]*domain.Payment, error)
	CapturePaymentFunc          func(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error)
	CreateRefundFunc            func(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error)
	GetRefundFunc               func(ctx context.Context, req *domain.GetRefundRequest) (*domain.Refund, error)
	ListRefundsFunc             func(ctx context.Context, req *domain.ListRefundsRequest) ([]*domain.Refund, error)
	GetInstrumentFunc           func(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error)
	ListInstrumentsFunc         func(ctx context.Context, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error)
	DeleteInstrumentFunc        func(ctx context.Context, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error)
	CreatePaymentLinkFunc       func(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error)
	GetPaymentLinkFunc          func(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error)
	CancelPaymentLinkFunc       func(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error)
	CreatePlanFunc              func(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error)
	GetPlanFunc                 func(ctx context.Context, req *domain.GetPlanRequest) (*domain.Plan, error)
	CreateSubscriptionFunc      func(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error)
	GetSubscriptionFunc         func(ctx context.Context, req *domain.GetSubscriptionRequest) (*domain.Subscription, error)
	CancelSubscriptionFunc      func(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error)
	PauseSubscriptionFunc       func(ctx context.Context, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error)
	ResumeSubscriptionFunc      func(ctx context.Context, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error)
	ChangePlanFunc              func(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error)
	GetSubscriptionPaymentsFunc func(ctx context.Context, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error)
	ChargeSubscriptionFunc      func(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error)

	// Call logs.
	CreateOrderCalls             []*domain.CreateOrderRequest
	GetOrderCalls                []*domain.GetOrderRequest
	CreatePaymentLinkCalls       []*domain.CreatePaymentLinkRequest
	CreateSubscriptionCalls      []*domain.CreateSubscriptionRequest
	GetSubscriptionCalls         []*domain.GetSubscriptionRequest
	CancelSubscriptionCalls      []*domain.CancelSubscriptionRequest
	PauseSubscriptionCalls       []*domain.PauseSubscriptionRequest
	ResumeSubscriptionCalls      []*domain.ResumeSubscriptionRequest
	ChangePlanCalls              []*domain.ChangePlanRequest
	GetSubscriptionPaymentsCalls []*domain.GetSubscriptionPaymentsRequest
	ChargeSubscriptionCalls      []*domain.ChargeSubscriptionRequest
	CreatePlanCalls              []*domain.CreatePlanRequest
	GetPlanCalls                 []*domain.GetPlanRequest
}

// NewFakeAdapter returns a FakeAdapter bound to provider (defaults to Cashfree when empty).
func NewFakeAdapter(provider domain.Provider) *FakeAdapter {
	if provider == "" {
		provider = domain.ProviderCashfree
	}
	return &FakeAdapter{providerName: provider}
}

// ---- OrderProvider ----

func (f *FakeAdapter) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	f.mu.Lock()
	f.CreateOrderCalls = append(f.CreateOrderCalls, req)
	fn := f.CreateOrderFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Order{}, nil
}

func (f *FakeAdapter) GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error) {
	f.mu.Lock()
	f.GetOrderCalls = append(f.GetOrderCalls, req)
	fn := f.GetOrderFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Order{}, nil
}

func (f *FakeAdapter) ListOrderPayments(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	f.mu.Lock()
	fn := f.ListOrderPaymentsFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return []*domain.Payment{}, nil
}

// ---- PaymentProvider ----

func (f *FakeAdapter) GetPayment(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	f.mu.Lock()
	fn := f.GetPaymentFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Payment{}, nil
}

func (f *FakeAdapter) ListPayments(ctx context.Context, req *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
	f.mu.Lock()
	fn := f.ListPaymentsFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return []*domain.Payment{}, nil
}

func (f *FakeAdapter) CapturePayment(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error) {
	f.mu.Lock()
	fn := f.CapturePaymentFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Payment{}, nil
}

// ---- RefundProvider ----

func (f *FakeAdapter) CreateRefund(ctx context.Context, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	f.mu.Lock()
	fn := f.CreateRefundFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Refund{}, nil
}

func (f *FakeAdapter) GetRefund(ctx context.Context, req *domain.GetRefundRequest) (*domain.Refund, error) {
	f.mu.Lock()
	fn := f.GetRefundFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Refund{}, nil
}

func (f *FakeAdapter) ListRefunds(ctx context.Context, req *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	f.mu.Lock()
	fn := f.ListRefundsFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return []*domain.Refund{}, nil
}

// ---- InstrumentProvider ----

func (f *FakeAdapter) GetInstrument(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	f.mu.Lock()
	fn := f.GetInstrumentFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Instrument{}, nil
}

func (f *FakeAdapter) ListInstruments(ctx context.Context, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	f.mu.Lock()
	fn := f.ListInstrumentsFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return []*domain.Instrument{}, nil
}

func (f *FakeAdapter) DeleteInstrument(ctx context.Context, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	f.mu.Lock()
	fn := f.DeleteInstrumentFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Instrument{}, nil
}

// ---- PaymentLinkProvider ----

func (f *FakeAdapter) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	f.mu.Lock()
	f.CreatePaymentLinkCalls = append(f.CreatePaymentLinkCalls, req)
	fn := f.CreatePaymentLinkFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.PaymentLink{}, nil
}

func (f *FakeAdapter) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	f.mu.Lock()
	fn := f.GetPaymentLinkFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.PaymentLink{}, nil
}

func (f *FakeAdapter) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	f.mu.Lock()
	fn := f.CancelPaymentLinkFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.PaymentLink{}, nil
}

// ---- PlanProvider ----

func (f *FakeAdapter) CreatePlan(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error) {
	f.mu.Lock()
	f.CreatePlanCalls = append(f.CreatePlanCalls, req)
	fn := f.CreatePlanFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Plan{}, nil
}

func (f *FakeAdapter) GetPlan(ctx context.Context, req *domain.GetPlanRequest) (*domain.Plan, error) {
	f.mu.Lock()
	f.GetPlanCalls = append(f.GetPlanCalls, req)
	fn := f.GetPlanFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Plan{}, nil
}

// ---- SubscriptionProvider ----

func (f *FakeAdapter) CreateSubscription(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	f.mu.Lock()
	f.CreateSubscriptionCalls = append(f.CreateSubscriptionCalls, req)
	fn := f.CreateSubscriptionFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *FakeAdapter) GetSubscription(ctx context.Context, req *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	f.mu.Lock()
	f.GetSubscriptionCalls = append(f.GetSubscriptionCalls, req)
	fn := f.GetSubscriptionFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *FakeAdapter) CancelSubscription(ctx context.Context, req *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	f.mu.Lock()
	f.CancelSubscriptionCalls = append(f.CancelSubscriptionCalls, req)
	fn := f.CancelSubscriptionFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *FakeAdapter) PauseSubscription(ctx context.Context, req *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	f.mu.Lock()
	f.PauseSubscriptionCalls = append(f.PauseSubscriptionCalls, req)
	fn := f.PauseSubscriptionFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *FakeAdapter) ResumeSubscription(ctx context.Context, req *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	f.mu.Lock()
	f.ResumeSubscriptionCalls = append(f.ResumeSubscriptionCalls, req)
	fn := f.ResumeSubscriptionFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *FakeAdapter) ChangePlan(ctx context.Context, req *domain.ChangePlanRequest) (*domain.Subscription, error) {
	f.mu.Lock()
	f.ChangePlanCalls = append(f.ChangePlanCalls, req)
	fn := f.ChangePlanFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.Subscription{}, nil
}

func (f *FakeAdapter) GetSubscriptionPayments(ctx context.Context, req *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	f.mu.Lock()
	f.GetSubscriptionPaymentsCalls = append(f.GetSubscriptionPaymentsCalls, req)
	fn := f.GetSubscriptionPaymentsFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return []*domain.SubscriptionPayment{}, nil
}

func (f *FakeAdapter) ChargeSubscription(ctx context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	f.mu.Lock()
	f.ChargeSubscriptionCalls = append(f.ChargeSubscriptionCalls, req)
	fn := f.ChargeSubscriptionFunc
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return &domain.SubscriptionPayment{}, nil
}

// ---- WebhookConsumerProvider ----

// VerifySignature always succeeds — the fake self-trusts its own payloads.
func (f *FakeAdapter) VerifySignature(ctx context.Context, payload []byte, headers map[string]string) error {
	return nil
}

// ParseEvent decodes the fake's WebhookEnvelope JSON into a canonical domain.WebhookEvent.
func (f *FakeAdapter) ParseEvent(ctx context.Context, payload []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	var env WebhookEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return nil, fmt.Errorf("fake ParseEvent: %w", err)
	}
	return &domain.WebhookEvent{
		Provider:        f.providerName,
		AccountID:       env.AccountID,
		EventType:       env.EventType,
		DedupeKey:       env.DedupeKey,
		RawVendorStatus: env.RawVendorStatus,
		Order:           env.Order,
		Payment:         env.Payment,
		Subscription:    env.Subscription,
		RawPayload:      payload,
	}, nil
}

// SupportedWebhookEvents lists the canonical events the fake can emit.
func (f *FakeAdapter) SupportedWebhookEvents() []domain.WebhookEventType {
	return []domain.WebhookEventType{
		domain.EventPaymentCaptured,
		domain.EventSubAuthenticated,
		domain.EventSubActivated,
		domain.EventSubCharged,
		domain.EventSubPaymentFailed,
		domain.EventSubOnHold,
		domain.EventSubCancelled,
		domain.EventSubCardExpiring,
		domain.EventSubCompleted,
		domain.EventSubExpired,
	}
}

// ---- MetadataMapper ----

func (f *FakeAdapter) MapOrderMetadata(ctx context.Context, metadata domain.Metadata) (map[string]any, error) {
	return map[string]any{}, nil
}

func (f *FakeAdapter) MapRefundMetadata(ctx context.Context, metadata domain.Metadata) (map[string]any, error) {
	return map[string]any{}, nil
}

func (f *FakeAdapter) MapPaymentLinkMetadata(ctx context.Context, metadata domain.Metadata) (map[string]any, error) {
	return map[string]any{}, nil
}

// ---- identity ----

// ProviderName returns the configured provider (default Cashfree).
func (f *FakeAdapter) ProviderName() domain.Provider {
	return f.providerName
}

// ProviderCapabilities returns an empty slice; the capability validator reads the
// hardcoded support matrix keyed by provider name, not this method.
func (f *FakeAdapter) ProviderCapabilities() []capabilities.Capability {
	return []capabilities.Capability{}
}

// WebhookEnvelope is the fake's self-describing webhook payload. The Harness marshals
// it to JSON and delivers it through the real WebhookService.HandleEvent; ParseEvent
// unmarshals it back into a domain.WebhookEvent.
type WebhookEnvelope struct {
	Provider        domain.Provider         `json:"provider"`
	AccountID       string                  `json:"account_id,omitempty"`
	EventType       domain.WebhookEventType `json:"event_type"`
	DedupeKey       string                  `json:"dedupe_key"`
	RawVendorStatus string                  `json:"raw_vendor_status,omitempty"`
	Order           *domain.Order           `json:"order,omitempty"`
	Payment         *domain.Payment         `json:"payment,omitempty"`
	Subscription    *domain.Subscription    `json:"subscription,omitempty"`
}

// Marshal returns the JSON bytes for this envelope.
func (e *WebhookEnvelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
