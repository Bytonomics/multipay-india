package routing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/logging"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// apiResponse is the JSON envelope for all webhook handler responses.
type apiResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WebhookHandler implements http.Handler for webhook routing and processing.
// It provides the core webhook flow: endpoint matching, deduplication, signature verification,
// event parsing, and handler dispatch.
type WebhookHandler struct {
	matcher        *EndpointMatcher
	adapter        ports.ProviderAdapter
	store          ports.WebhookStore
	logger         ports.Logger
	handlers       map[domain.WebhookEventType]domain.WebhookEventHandler
	defaultHandler domain.WebhookEventHandler
}

// NewWebhookHandler creates a new WebhookHandler with the given dependencies.
// All dependencies are required (not optional).
// The logger is wrapped with CallerLogger to automatically capture caller information.
func NewWebhookHandler(
	matcher *EndpointMatcher,
	adapter ports.ProviderAdapter,
	store ports.WebhookStore,
	logger ports.Logger,
	defaultHandler domain.WebhookEventHandler,
) *WebhookHandler {
	if defaultHandler == nil {
		panic("defaultHandler is required (cannot be nil); all unregistered and unknown events route here")
	}
	if store == nil {
		panic("WebhookStore is required (cannot be nil); it is the durable capture of received events")
	}
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	// Wrap the logger to automatically capture caller information
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &WebhookHandler{
		matcher:        matcher,
		adapter:        adapter,
		store:          store,
		logger:         wrappedLogger,
		handlers:       make(map[domain.WebhookEventType]domain.WebhookEventHandler),
		defaultHandler: defaultHandler,
	}
}

// RegisterEventHandler registers a handler for a specific event type.
// When a webhook event of this type is received and successfully processed,
// this handler will be called with the parsed event.
// Calling RegisterEventHandler multiple times for the same eventType replaces the previous handler.
func (h *WebhookHandler) RegisterEventHandler(eventType domain.WebhookEventType, handler domain.WebhookEventHandler) {
	h.handlers[eventType] = handler
}

// ServeHTTP implements the http.Handler interface, processing incoming webhook requests.
// It executes the 8-step webhook flow:
//
// 1. Extract provider + accountID from request path via EndpointMatcher
// 2. Read and store raw request body (durable capture)
// 3. Check for duplicates using SHA256(body) as dedupeKey
// 4. If duplicate, return 200 ACK immediately
// 5. Validate that the webhook provider matches the configured adapter provider
// 6. Verify webhook signature using adapter.VerifySignature()
// 7. Parse event (best-effort; parse failure does NOT abort dispatch)
// 8. Dispatch to registered handler or DefaultHandler (typed dispatch); always 2xx after sig verify
//
// HTTP responses follow this pattern:
// - 200: Success, duplicate, or after signature verification (always 2xx after sig verify)
// - 400: Bad request (invalid path, provider mismatch, signature verification failed)
// - 500: Other infrastructure errors
//
// All responses are JSON with "code" and "message" fields.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		if err := respondError(ctx, h.logger, w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is accepted"); err != nil {
			h.logger.Error(ctx, "failed to send method not allowed response", "error", err.Error())
		}
		return
	}

	// 1. Match endpoint to extract provider and accountID from path
	provider, accountID, ok := h.matcher.Match(r.URL.Path)
	if !ok {
		if err := respondError(ctx, h.logger, w, http.StatusBadRequest, "INVALID_ENDPOINT", "Invalid webhook endpoint"); err != nil {
			// Response write failed; connection likely broken
			return
		}
		return
	}

	// 2. Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if respErr := respondError(ctx, h.logger, w, http.StatusBadRequest, "READ_BODY_FAILED", "Failed to read request body"); respErr != nil {
			h.logger.Error(ctx, "failed to send error response after body read failure", "error", respErr.Error())
		}
		return
	}
	defer r.Body.Close()

	// 3. Compute dedupeKey early (needed for both IsDuplicate and MarkProcessed)
	dedupeKey := dedupeHash(body)

	// Store raw payload for audit/recovery (G8: mandatory, no nil guard)
	if storeErr := h.store.StoreRawPayload(ctx, provider, accountID, body); storeErr != nil {
		// Log but don't fail - storage is best-effort
		if respErr := respondError(ctx, h.logger, w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store webhook payload"); respErr != nil {
			h.logger.Error(ctx, "failed to send storage error response", "error", respErr.Error())
		}
		return
	}

	// 4. Check for duplicate via dedupeKey
	isDuplicate, dedupErr := h.store.IsDuplicate(ctx, provider, accountID, dedupeKey)
	if dedupErr != nil {
		// Log but don't fail - dedup check is best-effort
		if respErr := respondError(ctx, h.logger, w, http.StatusInternalServerError, "DEDUP_ERROR", "Failed to check for duplicates"); respErr != nil {
			h.logger.Error(ctx, "failed to send dedup error response", "error", respErr.Error())
		}
		return
	}

	if isDuplicate {
		// Duplicate webhook - return 200 ACK immediately (idempotent)
		if err := respondSuccess(ctx, h.logger, w, "DUPLICATE_ACK", "Duplicate webhook, acknowledged"); err != nil {
			h.logger.Error(ctx, "failed to send duplicate ack response", "error", err.Error())
		}
		return
	}

	// 5. Use configured adapter (validate provider matches)
	adapterProvider := h.adapter.ProviderName()
	if provider != adapterProvider {
		if respErr := respondError(ctx, h.logger, w, http.StatusBadRequest, "PROVIDER_MISMATCH", fmt.Sprintf("Webhook provider %s does not match configured provider %s", provider, adapterProvider)); respErr != nil {
			h.logger.Error(ctx, "failed to send provider mismatch response", "error", respErr.Error())
		}
		return
	}

	// Extract headers from request
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 6. Verify signature
	if verifyErr := h.adapter.VerifySignature(ctx, body, headers); verifyErr != nil {
		if errors.Is(verifyErr, domain.ErrWebhookVerificationFailed) {
			if err := respondError(ctx, h.logger, w, http.StatusBadRequest, "SIGNATURE_INVALID", "Webhook signature verification failed"); err != nil {
				h.logger.Error(ctx, "failed to send signature invalid response", "error", err.Error())
			}
		} else {
			if err := respondError(ctx, h.logger, w, http.StatusInternalServerError, "VERIFICATION_ERROR", "Webhook verification error"); err != nil {
				h.logger.Error(ctx, "failed to send verification error response", "error", err.Error())
			}
		}
		return
	}

	// 7. Parse event — best-effort; parse failure does NOT abort dispatch
	event, parseErr := h.adapter.ParseEvent(ctx, body, headers)
	if event == nil {
		event = &domain.WebhookEvent{}
	}
	// D11: always populate raw/source fields regardless of parse outcome
	event.Provider = provider
	event.AccountID = accountID
	event.WebhookURL = r.URL.Path
	event.RawPayload = body
	if event.RawHeaders == nil {
		event.RawHeaders = headers
	}
	if parseErr != nil {
		event.ParseError = parseErr.Error()
		if event.EventType == "" {
			event.EventType = domain.EventUnknown
		}
	}

	// 8. Dispatch — specific handler or DefaultHandler; never 202
	handler, ok := h.handlers[event.EventType]
	if !ok {
		handler = h.defaultHandler
	}
	if err := handler(ctx, event); err != nil {
		// G8: handler error → log at error, respond 2xx. Event is already persisted by StoreRawPayload
		// and NOT MarkProcessed, so it's queryable as "received but unprocessed" for replay.
		// Never let a buggy handler disable the endpoint (vendors auto-disable on repeated 5xx).
		h.logger.Error(ctx, "webhook handler returned error", "error", err.Error(), "eventType", string(event.EventType))
		// fall through to 2xx — DO NOT call MarkProcessed
	} else {
		// Success: mark as processed to prevent reprocessing
		if markErr := h.store.MarkProcessed(ctx, provider, accountID, dedupeKey); markErr != nil {
			h.logger.Error(ctx, "failed to mark webhook as processed", "error", markErr.Error())
		}
	}

	// D12: always respond 2xx after signature verification — prevents vendor auto-disabling on handler errors
	if err := respondSuccess(ctx, h.logger, w, "ACK", "Webhook processed"); err != nil {
		h.logger.Error(ctx, "failed to send success response", "error", err.Error())
	}
}

// dedupeHash computes a SHA256 hash of the body as a hex string.
// Used to uniquely identify duplicate webhook payloads.
func dedupeHash(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

// respondError sends a JSON error response with the given HTTP status code.
// All error responses have "code" and "message" fields.
// Returns an error if JSON encoding fails (response may be partially written).
func respondError(ctx context.Context, logger ports.Logger, w http.ResponseWriter, statusCode int, errCode, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(apiResponse{Code: errCode, Message: message}); err != nil {
		logger.Error(ctx, "failed to encode error response in respondError", "error", err.Error(), "statusCode", statusCode, "errCode", errCode)

		return fmt.Errorf("encode error response: %w", err)
	}
	return nil
}

// respondSuccess sends a JSON success response (HTTP 200 OK).
// Response has "code" and "message" fields.
// Returns an error if JSON encoding fails (response may be partially written).
func respondSuccess(ctx context.Context, logger ports.Logger, w http.ResponseWriter, code, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiResponse{Code: code, Message: message}); err != nil {
		logger.Error(ctx, "failed to encode success response in respondSuccess", "error", err.Error(), "code", code)

		return fmt.Errorf("encode success response: %w", err)
	}
	return nil
}
