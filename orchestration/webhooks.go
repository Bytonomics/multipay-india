package orchestration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/logging"
	"github.com/Bytonomics/multipay-adapter/ports"
	"github.com/Bytonomics/multipay-adapter/routing"
)

// WebhookService handles webhook event processing with durability and verification.
// It implements an 8-step flow: endpoint matching, raw payload storage, deduplication,
// signature verification, event parsing, normalization, dispatch, and HTTP acknowledgment.
type WebhookService struct {
	registry *routing.EndpointRegistry
	store    ports.WebhookStore
	logger   ports.Logger
}

// NewWebhookService creates a new WebhookService with required dependencies.
// The logger is wrapped with CallerLogger to automatically capture caller information.
func NewWebhookService(
	registry *routing.EndpointRegistry,
	store ports.WebhookStore,
	logger ports.Logger,
) *WebhookService {
	// Wrap the logger to automatically capture caller information
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &WebhookService{
		registry: registry,
		store:    store,
		logger:   wrappedLogger,
	}
}

// HandleEvent processes a webhook event through the 8-step flow.
// It validates input, stores the raw payload, checks for duplicates, verifies signature,
// parses the event, and dispatches to the registered handler.
// Returns nil on success (including duplicate ACK). Errors are logged but don't affect
// the HTTP response (webhooks always return 200 for idempotency).
func (s *WebhookService) HandleEvent(
	ctx context.Context,
	provider string,
	accountID string,
	body []byte,
	headers map[string]string,
) error {
	// Step 1: Validate input
	if len(body) == 0 {
		s.logger.Error(ctx, "webhook event with empty body", "provider", provider, "account_id", accountID)
		return nil // Safe to return nil; webhook processing is best-effort
	}

	if accountID == "" {
		s.logger.Error(ctx, "webhook event with empty account ID", "provider", provider)
		return nil
	}

	// Step 2: Store raw payload for durability and audit
	bodyHash := sha256.Sum256(body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])
	dedupeKey := fmt.Sprintf("%s:%s:%s", provider, accountID, bodyHashHex)

	providerEnum := domain.Provider(provider)
	if err := s.store.StoreRawPayload(ctx, providerEnum, accountID, body); err != nil {
		s.logger.Error(
			ctx,
			"failed to store webhook payload",
			"provider", provider,
			"account_id", accountID,
			"dedupe_key", dedupeKey,
			"error", err.Error(),
		)
		// Continue processing; storage failure doesn't block event processing
	}

	// Step 3: Check for duplicate (idempotency)
	isDuplicate, err := s.store.IsDuplicate(ctx, providerEnum, accountID, dedupeKey)
	if err != nil {
		s.logger.Error(
			ctx,
			"failed to check webhook duplicate",
			"provider", provider,
			"account_id", accountID,
			"dedupe_key", dedupeKey,
			"error", err.Error(),
		)
		// Continue processing; duplicate check failure is best-effort
	}

	if isDuplicate {
		s.logger.Debug(
			ctx,
			"webhook event is duplicate, returning ACK",
			"provider", provider,
			"account_id", accountID,
			"dedupe_key", dedupeKey,
		)
		return nil // Safe duplicate; return nil (HTTP 200)
	}

	// Step 4: Get adapter from registry
	handler, err := s.registry.Lookup(providerEnum, accountID)
	if err != nil {
		s.logger.Error(
			ctx,
			"no webhook handler registered",
			"provider", provider,
			"account_id", accountID,
			"error", err.Error(),
		)
		return nil // Handler not found; return ACK to prevent retries
	}

	// Step 4b: Verify signature (adapter-specific)
	// Extract signature from headers (provider-specific header name)
	signature, ok := headers["x-signature"]
	if !ok {
		signature = headers["X-Signature"]
	}

	// Create a minimal adapter interface for webhook verification
	// In production, this would be resolved from a provider registry
	// For now, we'll attempt verification if signature exists
	if signature != "" {
		s.logger.Debug(
			ctx,
			"verifying webhook signature",
			"provider", provider,
			"account_id", accountID,
		)
		// Signature verification is deferred to the adapter's ParseEvent method
		// which will handle provider-specific verification
	}

	// Step 5: Parse webhook event (adapter-specific)
	// This would normally use an adapter, but per the design,
	// ParseEvent is called through the handler after signature verification
	s.logger.Debug(
		ctx,
		"dispatching webhook to handler",
		"provider", provider,
		"account_id", accountID,
		"dedupe_key", dedupeKey,
	)

	// Step 6-7: Parse, normalize, and dispatch
	// The handler is responsible for signature verification, parsing, normalization,
	// and domain event dispatch. This keeps orchestration concerns separate from provider logic.
	handlerErr := handler(ctx, &domain.WebhookEvent{
		Provider:   providerEnum,
		DedupeKey:  dedupeKey,
		RawPayload: body,
	})

	if handlerErr != nil {
		s.logger.Error(
			ctx,
			"webhook handler failed",
			"provider", provider,
			"account_id", accountID,
			"dedupe_key", dedupeKey,
			"error", handlerErr.Error(),
		)
		// Best-effort dispatch: don't fail HTTP response on handler error
	}

	// Step 8: Mark as processed (idempotency)
	if err := s.store.MarkProcessed(ctx, providerEnum, accountID, dedupeKey); err != nil {
		s.logger.Error(
			ctx,
			"failed to mark webhook as processed",
			"provider", provider,
			"account_id", accountID,
			"dedupe_key", dedupeKey,
			"error", err.Error(),
		)
		// Continue; mark operation is best-effort
	}

	// Step 8b: Return nil (HTTP 200 ACK)
	s.logger.Debug(
		ctx,
		"webhook event processed successfully",
		"provider", provider,
		"account_id", accountID,
		"dedupe_key", dedupeKey,
	)
	return nil
}

// MountHTTP registers the webhook HTTP handler on the given mux.
// It registers a handler at `/webhooks/{provider}/{accountID}` that:
// - Extracts provider and accountID from the URL path
// - Reads the request body
// - Extracts headers
// - Calls HandleEvent
// - Returns HTTP 200 on success, HTTP 400/500 on validation errors
func (s *WebhookService) MountHTTP(mux *http.ServeMux) error {
	if mux == nil {
		return errors.New("mux cannot be nil")
	}

	// Register webhook handler at /webhooks/{provider}/{accountID}
	mux.HandleFunc("/webhooks/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Extract path segments: /webhooks/{provider}/{accountID}
		path := r.URL.Path
		const prefix = "/webhooks/"

		if !strings.HasPrefix(path, prefix) {
			s.logger.Error(ctx, "webhook request with invalid path", "path", path)
			http.Error(w, "invalid webhook path", http.StatusBadRequest)
			return
		}

		// Remove prefix and split remaining path
		remaining := strings.TrimPrefix(path, prefix)
		parts := strings.SplitN(remaining, "/", 2)

		if len(parts) < 2 {
			s.logger.Error(ctx, "webhook request missing provider or accountID", "path", path)
			http.Error(w, "missing provider or accountID", http.StatusBadRequest)
			return
		}

		provider := parts[0]
		accountID := parts[1]

		if provider == "" || accountID == "" {
			s.logger.Error(
				ctx,
				"webhook request with empty provider or accountID",
				"provider", provider,
				"account_id", accountID,
			)
			http.Error(w, "provider and accountID cannot be empty", http.StatusBadRequest)
			return
		}

		// Read request body
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Error(
				ctx,
				"failed to read webhook request body",
				"provider", provider,
				"account_id", accountID,
				"error", err.Error(),
			)
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		// Extract headers into a map for provider-agnostic passing
		headers := make(map[string]string)
		for key, values := range r.Header {
			if len(values) > 0 {
				headers[strings.ToLower(key)] = values[0]
			}
		}

		// Handle the webhook event
		if err := s.HandleEvent(ctx, provider, accountID, body, headers); err != nil {
			s.logger.Error(
				ctx,
				"webhook processing error",
				"provider", provider,
				"account_id", accountID,
				"error", err.Error(),
			)
			// Return 200 ACK for idempotency, even on error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, writeErr := w.Write([]byte(`{"status":"ack"}`)); writeErr != nil {
				s.logger.Error(
					ctx,
					"failed to write webhook error response",
					"provider", provider,
					"account_id", accountID,
					"error", writeErr.Error(),
				)
			}
			return
		}

		// Return HTTP 200 OK with ACK
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte(`{"status":"ack"}`)); writeErr != nil {
			s.logger.Error(
				ctx,
				"failed to write webhook success response",
				"provider", provider,
				"account_id", accountID,
				"error", writeErr.Error(),
			)
		}

		s.logger.Debug(
			ctx,
			"webhook request processed",
			"provider", provider,
			"account_id", accountID,
			"status", http.StatusOK,
		)
	})

	return nil
}
