package orchestration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/logging"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
	"github.com/Bytonomics/multipay-india/multipay-go/routing"
)

type WebhookService struct {
	adapter  ports.ProviderAdapter
	provider domain.Provider
	pipeline *hooks.Pipeline
	store    ports.WebhookStore
	registry *routing.EndpointRegistry
	handlers map[domain.WebhookEventType]domain.WebhookEventHandler
	logger   ports.Logger
}

func NewWebhookService(
	provider domain.Provider,
	adapter ports.ProviderAdapter,
	pipeline *hooks.Pipeline,
	store ports.WebhookStore,
	registry *routing.EndpointRegistry,
	logger ports.Logger,
) *WebhookService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)
	return &WebhookService{
		adapter:  adapter,
		provider: provider,
		pipeline: pipeline,
		store:    store,
		registry: registry,
		handlers: make(map[domain.WebhookEventType]domain.WebhookEventHandler),
		logger:   wrappedLogger,
	}
}

// RegisterHandler registers an event handler for a specific webhook event type.
func (s *WebhookService) RegisterHandler(eventType domain.WebhookEventType, handler domain.WebhookEventHandler) {
	s.handlers[eventType] = handler
}

// HandleEvent implements the 8-step webhook handling flow:
// 1. Resolve adapter from provider
// 2. Store raw payload (best-effort)
// 3. Verify signature
// 4. Parse event
// 5. Check for duplicate
// 6. Execute before-hooks
// 7. Dispatch to registered handler
// 8. Mark processed
func (s *WebhookService) HandleEvent(ctx context.Context, provider domain.Provider, accountID string, payload []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	// Step 1: Validate that the provider matches the configured adapter
	if provider != s.provider {
		return nil, fmt.Errorf("webhook provider %s does not match client provider %s: %w", provider, s.provider, domain.ErrProviderNotFound)
	}
	adapter := s.adapter

	// Step 2: Store raw payload (ledger-first, best-effort)
	if s.store != nil {
		if storeErr := s.store.StoreRawPayload(ctx, provider, accountID, payload); storeErr != nil {
			s.logger.Error(ctx, "failed to store raw webhook payload", "error", storeErr.Error(), "provider", string(provider))
		}
	}

	// Step 3: Verify signature
	if verifyErr := adapter.VerifySignature(ctx, payload, headers); verifyErr != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", verifyErr)
	}

	// Step 4: Parse event
	event, err := adapter.ParseEvent(ctx, payload, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}
	if event == nil {
		return nil, errors.New("webhook event parsing returned nil")
	}

	// Step 5: Check for duplicate
	if s.store != nil {
		isDuplicate, err := s.store.IsDuplicate(ctx, provider, accountID, event.DedupeKey)
		if err != nil {
			s.logger.Error(ctx, "failed to check webhook duplicate", "error", err.Error())
		}
		if isDuplicate {
			s.logger.Info(ctx, "webhook event is duplicate, skipping handler dispatch")
			return event, nil
		}
	}

	// Step 6: Execute before-hooks
	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "WebhookEvent",
		RequestData: event,
		StartTime:   time.Now(),
	}
	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("webhook before-hook execution failed: %w", hookErr)
	}

	// Step 7: Dispatch to registered handler
	if handler, exists := s.handlers[event.EventType]; exists {
		if err := handler(ctx, event); err != nil {
			if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
				s.logger.Error(ctx, "failed to execute error hook", "error", hookErr.Error())
			}
			return nil, fmt.Errorf("webhook event handler failed: %w", err)
		}
	} else {
		s.logger.Debug(ctx, "no handler registered for webhook event type", "eventType", string(event.EventType))
	}

	// Step 8: Mark processed
	if s.store != nil {
		if err := s.store.MarkProcessed(ctx, provider, accountID, event.DedupeKey); err != nil {
			s.logger.Error(ctx, "failed to mark webhook as processed", "error", err.Error(), "provider", string(provider), "accountID", accountID)
		}
	}

	// Execute after-hooks
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		s.logger.Error(ctx, "failed to execute after-hook", "error", err.Error())
	}

	return event, nil
}

// Handler returns the webhook endpoint as a framework-agnostic http.Handler. This is the single,
// portable way to mount webhook handling on ANY Go HTTP router — net/http, chi, Echo (via
// echo.WrapHandler), gin (via gin.WrapH), Fiber, etc. Because every router accepts an http.Handler,
// the library never needs a per-framework mount method. It builds the endpoint matcher, the mandatory
// non-nil default handler (returns nil => 2xx), constructs the routing.WebhookHandler, and registers
// all currently-registered event handlers.
//
// IMPORTANT: all RegisterHandler calls MUST happen BEFORE Handler is called — the handlers map is
// snapshotted into the returned handler at construction time.
//
// The EndpointMatcher re-parses the FULL request path ({basePath}/{provider}/{accountID}), so the
// consumer MUST mount this on a PREFIX/subtree route that does NOT strip basePath. A bare exact-path
// mount (e.g. net/http mux.Handle(basePath, h) without a trailing slash) matches ONLY basePath and
// would 404 the real webhook URL — always mount the subtree:
//   - net/http: mux.Handle(basePath+"/", svc.Handler(basePath))         // trailing slash => subtree
//   - chi:      r.Handle(basePath+"/*", svc.Handler(basePath))
//   - Echo:     e.Any(basePath+"/*", echo.WrapHandler(svc.Handler(basePath)))
//   - gin:      r.Any(basePath+"/*path", gin.WrapH(svc.Handler(basePath)))
func (s *WebhookService) Handler(basePath string) http.Handler {
	matcher := routing.NewEndpointMatcher(basePath)
	defaultHandler := func(ctx context.Context, ev *domain.WebhookEvent) error {
		s.logger.Info(ctx, "webhook event received (no specific handler)", "eventType", string(ev.EventType))
		return nil
	}
	h := routing.NewWebhookHandler(matcher, s.adapter, s.store, s.logger, defaultHandler)
	for evType, handler := range s.handlers {
		h.RegisterEventHandler(evType, handler)
	}
	return h
}
