package fake

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/client"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// WebhookEmitter is the injectable seam for emitting webhook events into the
// library. The default implementation is *Harness, which drives the REAL
// WebhookService.HandleEvent via the configured Scheduler after a wall-clock delay.
type WebhookEmitter interface {
	EmitWebhookAfter(env *WebhookEnvelope)
	WaitForWebhooks()
}

// compile-time interface check
var _ WebhookEmitter = (*Harness)(nil)

// Config configures a Harness. Every dependency is an interface with a default,
// so any of Store, Clock, Logger, Scheduler can be swapped in a test.
type Config struct {
	Provider            domain.Provider                           // default: domain.ProviderCashfree
	Store               ports.WebhookStore                        // default: *InMemoryWebhookStore
	Clock               ports.Clock                               // default: ports.NewRealClock()
	Logger              ports.Logger                              // default: ports.NewNoopLogger()
	Scheduler           Scheduler                                 // default: *WallClockScheduler
	AccountID           string                                    // default: "test"
	WebhookDelays       map[domain.WebhookEventType]time.Duration // per-event wall-clock delay
	DefaultWebhookDelay time.Duration                             // fallback delay when not in WebhookDelays
}

// Harness wires a real client.MultiPayClient around a FakeAdapter so ALL library
// flows run through the genuine orchestration (validation, capability checks, hook
// pipeline) without hitting a vendor. It also emits async webhooks with wall-clock delays.
type Harness struct {
	Client    *client.MultiPayClient
	Adapter   *FakeAdapter
	Store     ports.WebhookStore
	Clock     ports.Clock
	Scheduler Scheduler

	logger       ports.Logger
	provider     domain.Provider
	accountID    string
	delays       map[domain.WebhookEventType]time.Duration
	defaultDelay time.Duration

	mu          sync.Mutex
	webhookErrs []error
}

// NewHarness builds a Harness from a REQUIRED, valid config. The config is
// mandatory: a nil config is a programming error and panics — callers must pass an
// explicit *Config (e.g. &Config{} for all-default test deps, or one with fields set).
// The config's optional test-infra dependencies (Provider, Store, Clock, Logger,
// Scheduler) default to their standard test implementations when left zero; Provider
// defaults to domain.ProviderCashfree so the hardcoded capability support matrix lines up.
func NewHarness(cfg *Config) *Harness {
	if cfg == nil {
		panic("fake.NewHarness: cfg is required (cannot be nil); pass a valid *Config such as &Config{} or &Config{Provider: domain.ProviderCashfree}")
	}
	provider := cfg.Provider
	if provider == "" {
		provider = domain.ProviderCashfree
	}
	store := cfg.Store
	if store == nil {
		store = NewInMemoryWebhookStore()
	}
	clk := cfg.Clock
	if clk == nil {
		clk = ports.NewRealClock()
	}
	logger := cfg.Logger
	if logger == nil {
		logger = ports.NewNoopLogger()
	}
	sched := cfg.Scheduler
	if sched == nil {
		sched = NewWallClockScheduler()
	}
	accountID := cfg.AccountID
	if accountID == "" {
		accountID = "test"
	}

	adapter := NewFakeAdapter(provider)
	c, err := client.NewClient(&client.ClientConfig{
		Provider:     adapter,
		WebhookStore: store,
		Clock:        clk,
		Logger:       logger,
	})
	if err != nil {
		panic(fmt.Sprintf("fake.NewHarness: client.NewClient failed: %v", err))
	}

	return &Harness{
		Client:       c,
		Adapter:      adapter,
		Store:        store,
		Clock:        clk,
		Scheduler:    sched,
		logger:       logger,
		provider:     provider,
		accountID:    accountID,
		delays:       cfg.WebhookDelays,
		defaultDelay: cfg.DefaultWebhookDelay,
	}
}

func (h *Harness) delayFor(t domain.WebhookEventType) time.Duration {
	if h.delays != nil {
		if d, ok := h.delays[t]; ok {
			return d
		}
	}
	return h.defaultDelay
}

// EmitWebhookAfter schedules the envelope to be delivered through the REAL
// WebhookService.HandleEvent after this event type's configured wall-clock delay.
// It stamps the harness provider/account onto the envelope if unset.
func (h *Harness) EmitWebhookAfter(env *WebhookEnvelope) {
	env.Provider = h.provider
	if env.AccountID == "" {
		env.AccountID = h.accountID
	}
	payload, err := env.Marshal()
	if err != nil {
		panic(fmt.Sprintf("fake.Harness.EmitWebhookAfter: marshal envelope failed: %v", err))
	}
	accountID := env.AccountID
	delay := h.delayFor(env.EventType)
	h.Scheduler.Schedule(delay, func() {
		ctx := context.Background()
		_, herr := h.Client.Webhooks().HandleEvent(ctx, h.provider, accountID, payload, map[string]string{})
		if herr != nil {
			// The webhook endpoint always returns 2xx after signature verification, so a
			// delivery/handler error must not crash the emitter. Handle it: log it AND
			// record it so tests can assert whether delivery succeeded via WebhookErrors().
			h.logger.Error(ctx, "fake webhook delivery failed", "provider", string(h.provider), "account_id", accountID, "error", herr.Error())
			h.mu.Lock()
			h.webhookErrs = append(h.webhookErrs, herr)
			h.mu.Unlock()
		}
	})
}

// WaitForWebhooks blocks until every scheduled webhook emission has been delivered.
func (h *Harness) WaitForWebhooks() {
	h.Scheduler.Wait()
}

// WebhookErrors returns a copy of every error returned by HandleEvent for the
// webhooks emitted so far. Call after WaitForWebhooks(). An empty result means
// every emitted webhook was delivered without error.
func (h *Harness) WebhookErrors() []error {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]error, len(h.webhookErrs))
	copy(out, h.webhookErrs)
	return out
}
