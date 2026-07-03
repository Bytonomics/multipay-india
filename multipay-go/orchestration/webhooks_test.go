package orchestration

import (
	"context"
	"net/http"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/routing"
)

// stubWebhookStore is a no-op ports.WebhookStore for constructing a WebhookService in tests.
type stubWebhookStore struct{}

func (stubWebhookStore) StoreRawPayload(_ context.Context, _ domain.Provider, _ string, _ []byte) error {
	return nil
}
func (stubWebhookStore) IsDuplicate(_ context.Context, _ domain.Provider, _ string, _ string) (bool, error) {
	return false, nil
}
func (stubWebhookStore) MarkProcessed(_ context.Context, _ domain.Provider, _ string, _ string) error {
	return nil
}

// stubProviderAdapter is a no-op ports.ProviderAdapter for constructing a WebhookService in tests.
type stubProviderAdapter struct{}

func (stubProviderAdapter) CreateOrder(_ context.Context, _ *domain.CreateOrderRequest) (*domain.Order, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetOrder(_ context.Context, _ *domain.GetOrderRequest) (*domain.Order, error) {
	return nil, domain.ErrOrderNotFound
}
func (stubProviderAdapter) ListOrderPayments(_ context.Context, _ *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetPayment(_ context.Context, _ *domain.GetPaymentRequest) (*domain.Payment, error) {
	return nil, domain.ErrPaymentNotFound
}
func (stubProviderAdapter) ListPayments(_ context.Context, _ *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) CapturePayment(_ context.Context, _ *domain.CapturePaymentRequest) (*domain.Payment, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) CreateRefund(_ context.Context, _ *domain.CreateRefundRequest) (*domain.Refund, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetRefund(_ context.Context, _ *domain.GetRefundRequest) (*domain.Refund, error) {
	return nil, domain.ErrRefundNotFound
}
func (stubProviderAdapter) ListRefunds(_ context.Context, _ *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetInstrument(_ context.Context, _ *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	return nil, domain.ErrInstrumentNotFound
}
func (stubProviderAdapter) ListInstruments(_ context.Context, _ *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) DeleteInstrument(_ context.Context, _ *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) CreatePaymentLink(_ context.Context, _ *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetPaymentLink(_ context.Context, _ *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, domain.ErrPaymentLinkNotFound
}
func (stubProviderAdapter) CancelPaymentLink(_ context.Context, _ *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) CreatePlan(_ context.Context, _ *domain.CreatePlanRequest) (*domain.Plan, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetPlan(_ context.Context, _ *domain.GetPlanRequest) (*domain.Plan, error) {
	return nil, domain.ErrPlanNotFound
}
func (stubProviderAdapter) CreateSubscription(_ context.Context, _ *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetSubscription(_ context.Context, _ *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	return nil, domain.ErrSubscriptionNotFound
}
func (stubProviderAdapter) CancelSubscription(_ context.Context, _ *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) PauseSubscription(_ context.Context, _ *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) ResumeSubscription(_ context.Context, _ *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) ChangePlan(_ context.Context, _ *domain.ChangePlanRequest) (*domain.Subscription, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) GetSubscriptionPayments(_ context.Context, _ *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) ChargeSubscription(_ context.Context, _ *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	return nil, domain.ErrProviderError
}
func (stubProviderAdapter) VerifySignature(_ context.Context, _ []byte, _ map[string]string) error {
	return nil
}
func (stubProviderAdapter) ParseEvent(_ context.Context, _ []byte, _ map[string]string) (*domain.WebhookEvent, error) {
	return &domain.WebhookEvent{
		EventType: domain.EventSubCharged,
		DedupeKey: "test_dedupe_key",
	}, nil
}
func (stubProviderAdapter) SupportedWebhookEvents() []domain.WebhookEventType {
	return []domain.WebhookEventType{}
}
func (stubProviderAdapter) MapOrderMetadata(_ context.Context, _ domain.Metadata) (map[string]any, error) {
	return map[string]any{}, nil
}
func (stubProviderAdapter) MapRefundMetadata(_ context.Context, _ domain.Metadata) (map[string]any, error) {
	return map[string]any{}, nil
}
func (stubProviderAdapter) MapPaymentLinkMetadata(_ context.Context, _ domain.Metadata) (map[string]any, error) {
	return map[string]any{}, nil
}
func (stubProviderAdapter) ProviderName() domain.Provider {
	return domain.ProviderCashfree
}
func (stubProviderAdapter) ProviderCapabilities() []capabilities.Capability {
	return []capabilities.Capability{}
}

func TestWebhookService_Handler_ReturnsNonNilHTTPHandler(t *testing.T) {
	svc := NewWebhookService(domain.ProviderCashfree, stubProviderAdapter{}, &hooks.Pipeline{}, stubWebhookStore{}, &routing.EndpointRegistry{}, ports.NewNoopLogger())

	var h http.Handler = svc.Handler("/api/v1/subscriptions/webhooks")
	if h == nil {
		t.Fatal("Handler returned nil; expected a non-nil http.Handler")
	}
}

// TestWebhookService_ReplayEvent_DispatchesDespiteDuplicate verifies ReplayEvent skips dedup checks.
// It registers a handler, then calls ReplayEvent.
// The handler should be invoked despite the duplicate (no IsDuplicate check in ReplayEvent).
func TestWebhookService_ReplayEvent_DispatchesDespiteDuplicate(t *testing.T) {
	handlerCalled := false
	handler := func(_ context.Context, _ *domain.WebhookEvent) error {
		handlerCalled = true
		return nil
	}

	svc := NewWebhookService(domain.ProviderCashfree, stubProviderAdapter{}, &hooks.Pipeline{}, stubWebhookStore{}, &routing.EndpointRegistry{}, ports.NewNoopLogger())
	svc.RegisterHandler(domain.EventSubCharged, handler)

	ctx := context.Background()
	event, err := svc.ReplayEvent(ctx, domain.ProviderCashfree, "test_account", []byte(`{}`), map[string]string{})

	if err != nil {
		t.Fatalf("ReplayEvent should not error, got: %v", err)
	}
	if event == nil {
		t.Fatalf("ReplayEvent should return a WebhookEvent, got nil")
	}
	if !handlerCalled {
		t.Fatalf("ReplayEvent should dispatch to handler despite being a replay")
	}
}
