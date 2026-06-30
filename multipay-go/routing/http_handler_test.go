package routing

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// fakeAdapter implements ports.ProviderAdapter for webhook dispatch tests.
// It provides a minimal implementation with stubbed methods.
type fakeAdapter struct {
	provider       domain.Provider
	parseEventFunc func(ctx context.Context, body []byte, headers map[string]string) (*domain.WebhookEvent, error)
}

func (f *fakeAdapter) ProviderName() domain.Provider {
	return f.provider
}

func (f *fakeAdapter) ChargeSubscription(_ context.Context, req *domain.ChargeSubscriptionRequest) (*domain.SubscriptionPayment, error) {
	return &domain.SubscriptionPayment{
		PaymentID:   req.PaymentRef,
		Status:      domain.SubPaymentStatusSuccess,
		AmountMinor: req.AmountMinor,
	}, nil
}

func (f *fakeAdapter) VerifySignature(_ context.Context, _ []byte, _ map[string]string) error {
	return nil
}

func (f *fakeAdapter) ParseEvent(ctx context.Context, body []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	if f.parseEventFunc != nil {
		return f.parseEventFunc(ctx, body, headers)
	}
	return &domain.WebhookEvent{EventType: domain.EventSubCharged}, nil
}

func (f *fakeAdapter) SupportedWebhookEvents() []domain.WebhookEventType {
	return nil
}

// OrderProvider methods
func (f *fakeAdapter) CreateOrder(_ context.Context, _ *domain.CreateOrderRequest) (*domain.Order, error) {
	return nil, nil
}

func (f *fakeAdapter) GetOrder(_ context.Context, _ *domain.GetOrderRequest) (*domain.Order, error) {
	return nil, nil
}

func (f *fakeAdapter) ListOrderPayments(_ context.Context, _ *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	return nil, nil
}

// PaymentProvider methods
func (f *fakeAdapter) GetPayment(_ context.Context, _ *domain.GetPaymentRequest) (*domain.Payment, error) {
	return nil, nil
}

func (f *fakeAdapter) ListPayments(_ context.Context, _ *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
	return nil, nil
}

func (f *fakeAdapter) CapturePayment(_ context.Context, _ *domain.CapturePaymentRequest) (*domain.Payment, error) {
	return nil, nil
}

// RefundProvider methods
func (f *fakeAdapter) CreateRefund(_ context.Context, _ *domain.CreateRefundRequest) (*domain.Refund, error) {
	return nil, nil
}

func (f *fakeAdapter) GetRefund(_ context.Context, _ *domain.GetRefundRequest) (*domain.Refund, error) {
	return nil, nil
}

func (f *fakeAdapter) ListRefunds(_ context.Context, _ *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	return nil, nil
}

// InstrumentProvider methods
func (f *fakeAdapter) GetInstrument(_ context.Context, _ *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	return nil, nil
}

func (f *fakeAdapter) ListInstruments(_ context.Context, _ *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	return nil, nil
}

func (f *fakeAdapter) DeleteInstrument(_ context.Context, _ *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	return nil, nil
}

// PaymentLinkProvider methods
func (f *fakeAdapter) CreatePaymentLink(_ context.Context, _ *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, nil
}

func (f *fakeAdapter) GetPaymentLink(_ context.Context, _ *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, nil
}

func (f *fakeAdapter) CancelPaymentLink(_ context.Context, _ *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	return nil, nil
}

// PlanProvider methods
func (f *fakeAdapter) CreatePlan(_ context.Context, _ *domain.CreatePlanRequest) (*domain.Plan, error) {
	return nil, nil
}

func (f *fakeAdapter) GetPlan(_ context.Context, _ *domain.GetPlanRequest) (*domain.Plan, error) {
	return nil, nil
}

// SubscriptionProvider methods
func (f *fakeAdapter) CreateSubscription(_ context.Context, _ *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) GetSubscription(_ context.Context, _ *domain.GetSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) CancelSubscription(_ context.Context, _ *domain.CancelSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) PauseSubscription(_ context.Context, _ *domain.PauseSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) ResumeSubscription(_ context.Context, _ *domain.ResumeSubscriptionRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) ChangePlan(_ context.Context, _ *domain.ChangePlanRequest) (*domain.Subscription, error) {
	return nil, nil
}

func (f *fakeAdapter) GetSubscriptionPayments(_ context.Context, _ *domain.GetSubscriptionPaymentsRequest) ([]*domain.SubscriptionPayment, error) {
	return nil, nil
}

// MetadataMapper methods
func (f *fakeAdapter) MapOrderMetadata(_ context.Context, _ domain.Metadata) (map[string]any, error) {
	return nil, nil
}

func (f *fakeAdapter) MapRefundMetadata(_ context.Context, _ domain.Metadata) (map[string]any, error) {
	return nil, nil
}

func (f *fakeAdapter) MapPaymentLinkMetadata(_ context.Context, _ domain.Metadata) (map[string]any, error) {
	return nil, nil
}

func (f *fakeAdapter) ProviderCapabilities() []capabilities.Capability {
	return nil
}

// fakeStore implements ports.WebhookStore for testing.
type fakeStore struct {
	stored    map[string][]byte
	processed map[string]bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		stored:    make(map[string][]byte),
		processed: make(map[string]bool),
	}
}

func (s *fakeStore) StoreRawPayload(_ context.Context, provider domain.Provider, accountID string, payload []byte) error {
	key := provider.String() + "/" + accountID
	s.stored[key] = payload
	return nil
}

func (s *fakeStore) IsDuplicate(_ context.Context, provider domain.Provider, accountID string, dedupeKey string) (bool, error) {
	key := provider.String() + "/" + accountID + "/" + dedupeKey
	return s.processed[key], nil
}

func (s *fakeStore) MarkProcessed(_ context.Context, provider domain.Provider, accountID string, dedupeKey string) error {
	key := provider.String() + "/" + accountID + "/" + dedupeKey
	s.processed[key] = true
	return nil
}

// fakeLogger implements ports.Logger for testing.
type fakeLogger struct {
	errors []string
}

func (l *fakeLogger) Info(_ context.Context, _ string, _ ...any) {}

func (l *fakeLogger) Error(_ context.Context, msg string, _ ...any) {
	l.errors = append(l.errors, msg)
}

func (l *fakeLogger) Debug(_ context.Context, _ string, _ ...any) {}

// newTestHandler creates a WebhookHandler for testing with a registered endpoint.
func newTestHandler(
	adapter ports.ProviderAdapter,
	store ports.WebhookStore,
	defaultHandler domain.WebhookEventHandler,
) *WebhookHandler {
	matcher := NewEndpointMatcher("/webhooks")
	return NewWebhookHandler(matcher, adapter, store, &fakeLogger{}, defaultHandler)
}

// makeWebhookRequest creates an HTTP POST request for webhook testing.
func makeWebhookRequest(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/webhooks/cashfree/acc1", strings.NewReader(body))
}

// TestWebhookHandler_TypedEnumDispatch verifies that a registered handler for a specific
// event type is called instead of the DefaultHandler.
func TestWebhookHandler_TypedEnumDispatch(t *testing.T) {
	specificCalled := false
	defaultCalled := false

	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		defaultCalled = true
		return nil
	})

	// Register a specific handler for EventSubCharged
	h.RegisterEventHandler(domain.EventSubCharged, func(_ context.Context, _ *domain.WebhookEvent) error {
		specificCalled = true
		return nil
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{"charged":true}`))

	if !specificCalled {
		t.Error("specific handler was not called")
	}
	if defaultCalled {
		t.Error("DefaultHandler should not be called when specific handler exists")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestWebhookHandler_UnregisteredEventRoutesToDefault verifies that unregistered event types
// are routed to the DefaultHandler and return 200 (not 202).
func TestWebhookHandler_UnregisteredEventRoutesToDefault(t *testing.T) {
	defaultCalled := false

	adapter := &fakeAdapter{
		provider: domain.ProviderCashfree,
		parseEventFunc: func(_ context.Context, _ []byte, _ map[string]string) (*domain.WebhookEvent, error) {
			return &domain.WebhookEvent{EventType: domain.EventSubActivated}, nil
		},
	}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		defaultCalled = true
		return nil
	})

	// NO RegisterEventHandler call for EventSubActivated

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{}`))

	if !defaultCalled {
		t.Error("DefaultHandler should be called for unregistered events")
	}
	if w.Code == http.StatusAccepted {
		t.Error("202 is removed; should never return 202")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestWebhookHandler_EventUnknownRoutesToDefault verifies that EventUnknown events
// (when parse succeeded but event type not recognized) route to DefaultHandler with 2xx response.
func TestWebhookHandler_EventUnknownRoutesToDefault(t *testing.T) {
	defaultCalled := false
	receivedEvent := (*domain.WebhookEvent)(nil)

	adapter := &fakeAdapter{
		provider: domain.ProviderCashfree,
		parseEventFunc: func(_ context.Context, _ []byte, _ map[string]string) (*domain.WebhookEvent, error) {
			return &domain.WebhookEvent{EventType: domain.EventUnknown}, nil
		},
	}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, e *domain.WebhookEvent) error {
		defaultCalled = true
		receivedEvent = e
		return nil
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{}`))

	if !defaultCalled {
		t.Error("DefaultHandler should be called for EventUnknown")
	}
	if receivedEvent == nil {
		t.Fatal("event should not be nil")
	}
	if receivedEvent.EventType != domain.EventUnknown {
		t.Errorf("expected EventUnknown, got %s", receivedEvent.EventType)
	}
	if w.Code < 200 || w.Code >= 300 {
		t.Errorf("expected 2xx, got %d", w.Code)
	}
}

// TestWebhookHandler_NilDefaultHandlerPanics verifies that a nil DefaultHandler
// causes a panic during construction.
func TestWebhookHandler_NilDefaultHandlerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil defaultHandler, got none")
		}
	}()

	matcher := NewEndpointMatcher("/webhooks")
	_ = NewWebhookHandler(matcher, &fakeAdapter{provider: domain.ProviderCashfree}, newFakeStore(), &fakeLogger{}, nil)
}

// TestWebhookHandler_NilStorePanics verifies that a nil WebhookStore
// causes a panic during construction.
func TestWebhookHandler_NilStorePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil store, got none")
		}
	}()

	matcher := NewEndpointMatcher("/webhooks")
	defaultHandler := func(_ context.Context, _ *domain.WebhookEvent) error { return nil }
	_ = NewWebhookHandler(matcher, &fakeAdapter{provider: domain.ProviderCashfree}, nil, &fakeLogger{}, defaultHandler)
}

// TestWebhookHandler_ParseErrorGracefulDegradation verifies that when ParseEvent returns an error:
// 1. Response is 2xx (not 4xx)
// 2. DefaultHandler is still called with the event
// 3. The event has ParseError set
func TestWebhookHandler_ParseErrorGracefulDegradation(t *testing.T) {
	defaultCalledWith := (*domain.WebhookEvent)(nil)

	adapter := &fakeAdapter{
		provider: domain.ProviderCashfree,
		parseEventFunc: func(_ context.Context, _ []byte, _ map[string]string) (*domain.WebhookEvent, error) {
			return nil, errors.New("simulated parse failure")
		},
	}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, e *domain.WebhookEvent) error {
		defaultCalledWith = e
		return nil
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{"corrupt":true}`))

	if w.Code < 200 || w.Code >= 300 {
		t.Errorf("expected 2xx, got %d", w.Code)
	}
	if defaultCalledWith == nil {
		t.Fatal("DefaultHandler should be called even on parse error")
	}
	if defaultCalledWith.ParseError == "" {
		t.Error("event.ParseError must be set on parse failure")
	}
	if defaultCalledWith.EventType != domain.EventUnknown {
		t.Errorf("expected EventUnknown, got %s", defaultCalledWith.EventType)
	}
}

// TestWebhookHandler_HandlerErrorIs2xx verifies that when a handler returns an error,
// the response is still 2xx (not 5xx), preventing vendor auto-disabling.
func TestWebhookHandler_HandlerErrorIs2xx(t *testing.T) {
	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		return nil
	})

	h.RegisterEventHandler(domain.EventSubCharged, func(_ context.Context, _ *domain.WebhookEvent) error {
		return errors.New("handler processing failed")
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{}`))

	if w.Code < 200 || w.Code >= 300 {
		t.Errorf("expected 2xx even on handler error, got %d", w.Code)
	}
}

// TestWebhookHandler_DefaultHandlerErrorIs2xx verifies that when DefaultHandler returns an error,
// the response is still 2xx.
func TestWebhookHandler_DefaultHandlerErrorIs2xx(t *testing.T) {
	adapter := &fakeAdapter{
		provider: domain.ProviderCashfree,
		parseEventFunc: func(_ context.Context, _ []byte, _ map[string]string) (*domain.WebhookEvent, error) {
			return &domain.WebhookEvent{EventType: domain.EventSubActivated}, nil
		},
	}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		return errors.New("default handler failed")
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{}`))

	if w.Code < 200 || w.Code >= 300 {
		t.Errorf("expected 2xx even when DefaultHandler errors, got %d", w.Code)
	}
}

// TestWebhookHandler_DuplicateAcknowledged verifies that duplicate webhooks (same body hash)
// are acknowledged with 200 without calling handlers.
func TestWebhookHandler_DuplicateAcknowledged(t *testing.T) {
	handlerCalled := false
	store := newFakeStore()
	adapter := &fakeAdapter{provider: domain.ProviderCashfree}

	h := newTestHandler(adapter, store, func(_ context.Context, _ *domain.WebhookEvent) error {
		handlerCalled = true
		return nil
	})

	body := `{"order_id":"12345"}`

	// First request
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/webhooks/cashfree/acc1", strings.NewReader(body))
	h.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("first request expected 200, got %d", w1.Code)
	}
	if !handlerCalled {
		t.Error("handler should be called on first request")
	}

	// Second request with same body (duplicate)
	handlerCalled = false
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/cashfree/acc1", strings.NewReader(body))
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("duplicate request expected 200, got %d", w2.Code)
	}
	if handlerCalled {
		t.Error("handler should not be called on duplicate request")
	}
}

// TestWebhookHandler_RawPayloadAndHeadersPopulated verifies that event.RawPayload and event.RawHeaders
// are populated correctly from the request.
func TestWebhookHandler_RawPayloadAndHeadersPopulated(t *testing.T) {
	receivedEvent := (*domain.WebhookEvent)(nil)

	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, e *domain.WebhookEvent) error {
		receivedEvent = e
		return nil
	})

	bodyStr := `{"test":"data"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/cashfree/acc1", strings.NewReader(bodyStr))
	req.Header.Set("X-Webhook-Signature", "test-signature")

	h.ServeHTTP(w, req)

	if receivedEvent == nil {
		t.Fatal("event should not be nil")
	}
	if !bytes.Equal(receivedEvent.RawPayload, []byte(bodyStr)) {
		t.Errorf("expected RawPayload=%s, got %s", bodyStr, string(receivedEvent.RawPayload))
	}
	if receivedEvent.RawHeaders == nil {
		t.Error("RawHeaders should not be nil")
	}
	if sig, ok := receivedEvent.RawHeaders["X-Webhook-Signature"]; !ok || sig != "test-signature" {
		t.Error("X-Webhook-Signature not found in RawHeaders")
	}
}

// TestWebhookHandler_ProviderAndAccountIDPopulated verifies that event.Provider and event.AccountID
// are set from the request path.
func TestWebhookHandler_ProviderAndAccountIDPopulated(t *testing.T) {
	receivedEvent := (*domain.WebhookEvent)(nil)

	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, e *domain.WebhookEvent) error {
		receivedEvent = e
		return nil
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, makeWebhookRequest(`{}`))

	if receivedEvent == nil {
		t.Fatal("event should not be nil")
	}
	if receivedEvent.Provider != domain.ProviderCashfree {
		t.Errorf("expected Provider=%s, got %s", domain.ProviderCashfree, receivedEvent.Provider)
	}
	if receivedEvent.AccountID != "acc1" {
		t.Errorf("expected AccountID=acc1, got %s", receivedEvent.AccountID)
	}
}

// TestWebhookHandler_InvalidPathRejected verifies that invalid webhook paths are rejected with 400.
func TestWebhookHandler_InvalidPathRejected(t *testing.T) {
	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/invalid/path/extra", strings.NewReader(`{}`))
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid path, got %d", w.Code)
	}
}

// TestWebhookHandler_ProviderMismatchRejected verifies that a webhook from a different provider
// than the configured adapter is rejected with 400.
func TestWebhookHandler_ProviderMismatchRejected(t *testing.T) {
	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/razorpay/acc1", strings.NewReader(`{}`))
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for provider mismatch, got %d", w.Code)
	}
}

// TestWebhookHandler_MethodNotAllowedRejected verifies that non-POST requests are rejected.
func TestWebhookHandler_MethodNotAllowedRejected(t *testing.T) {
	adapter := &fakeAdapter{provider: domain.ProviderCashfree}
	h := newTestHandler(adapter, newFakeStore(), func(_ context.Context, _ *domain.WebhookEvent) error {
		return nil
	})

	tests := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/webhooks/cashfree/acc1", strings.NewReader(`{}`))
		h.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 for %s, got %d", method, w.Code)
		}
	}
}
