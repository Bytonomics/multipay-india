package routing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Bytonomics/multipay-adapter/logging"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// WebhookHandler implements http.Handler for webhook routing and processing.
// It provides the core webhook flow: endpoint matching, deduplication, signature verification,
// event parsing, and handler dispatch.
type WebhookHandler struct {
	matcher  *EndpointMatcher
	store    ports.WebhookStore
	resolver *ports.ProviderRegistry
	logger   ports.Logger
	handlers map[string]WebhookEventHandler
}

// NewWebhookHandler creates a new WebhookHandler with the given dependencies.
// All dependencies are required (not optional).
// The logger is wrapped with CallerLogger to automatically capture caller information.
func NewWebhookHandler(
	matcher *EndpointMatcher,
	store ports.WebhookStore,
	resolver *ports.ProviderRegistry,
	logger ports.Logger,
) *WebhookHandler {
	// Wrap the logger to automatically capture caller information
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &WebhookHandler{
		matcher:  matcher,
		store:    store,
		resolver: resolver,
		logger:   wrappedLogger,
		handlers: make(map[string]WebhookEventHandler),
	}
}

// RegisterEventHandler registers a handler for a specific event type.
// When a webhook event of this type is received and successfully processed,
// this handler will be called with the parsed event.
// Calling RegisterEventHandler multiple times for the same eventType replaces the previous handler.
func (h *WebhookHandler) RegisterEventHandler(eventType string, handler WebhookEventHandler) {
	h.handlers[eventType] = handler
}

// ServeHTTP implements the http.Handler interface, processing incoming webhook requests.
// It executes the 8-step webhook flow:
//
// 1. Extract provider + accountID from request path via EndpointMatcher
// 2. Read and store raw request body
// 3. Check for duplicates using SHA256(body) as dedupeKey
// 4. If duplicate, return 200 ACK immediately
// 5. Resolve the ProviderAdapter for the provider
// 6. Verify webhook signature using adapter.VerifySignature()
// 7. Parse event using adapter.ParseEvent()
// 8. Dispatch to registered WebhookEventHandler and mark as processed
//
// HTTP responses follow this pattern:
// - 200: Success or duplicate (safe to retry)
// - 202: No handler registered for event type (graceful, no error)
// - 400: Bad request (invalid path, signature, parse error)
// - 500: Handler execution error
//
// All responses are JSON with "code" and "message" fields.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
			if h.logger != nil {
				h.logger.Error(ctx, "failed to send error response after body read failure", "error", respErr.Error())
			}
		}
		return
	}
	defer r.Body.Close()

	// 3. Store raw payload for audit/recovery
	if h.store != nil {
		// Use dedupeKey = SHA256(body) for uniqueness
		dedupeKey := dedupeHash(body)
		if storeErr := h.store.StoreRawPayload(ctx, provider, accountID, body); storeErr != nil {
			// Log but don't fail - storage is best-effort
			if respErr := respondError(ctx, h.logger, w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store webhook payload"); respErr != nil {
				if h.logger != nil {
					h.logger.Error(ctx, "failed to send storage error response", "error", respErr.Error())
				}
			}
			return
		}

		// 4. Check for duplicate via dedupeKey
		isDuplicate, dedupErr := h.store.IsDuplicate(ctx, provider, accountID, dedupeKey)
		if dedupErr != nil {
			// Log but don't fail - dedup check is best-effort
			if respErr := respondError(ctx, h.logger, w, http.StatusInternalServerError, "DEDUP_ERROR", "Failed to check for duplicates"); respErr != nil {
				if h.logger != nil {
					h.logger.Error(ctx, "failed to send dedup error response", "error", respErr.Error())
				}
			}
			return
		}

		if isDuplicate {
			// Duplicate webhook - return 200 ACK immediately (idempotent)
			if err := respondSuccess(ctx, h.logger, w, "DUPLICATE_ACK", "Duplicate webhook, acknowledged"); err != nil {
				if h.logger != nil {
					h.logger.Error(ctx, "failed to send duplicate ack response", "error", err.Error())
				}
			}
			return
		}
	}

	// 5. Resolve adapter for provider
	adapter, resolveErr := h.resolver.Resolve(ctx, provider)
	if resolveErr != nil {
		if respErr := respondError(ctx, h.logger, w, http.StatusBadRequest, "PROVIDER_NOT_FOUND", fmt.Sprintf("Provider %s not found", provider)); respErr != nil {
			if h.logger != nil {
				h.logger.Error(ctx, "failed to send provider not found response", "error", respErr.Error())
			}
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
	if verifyErr := adapter.VerifySignature(ctx, body, headers); verifyErr != nil {
		if respErr := respondError(ctx, h.logger, w, http.StatusBadRequest, "SIGNATURE_INVALID", "Webhook signature verification failed"); respErr != nil {
			if h.logger != nil {
				h.logger.Error(ctx, "failed to send signature invalid response", "error", respErr.Error())
			}
		}
		return
	}

	// 7. Parse event
	event, parseErr := adapter.ParseEvent(ctx, body, headers)
	if parseErr != nil {
		if respErr := respondError(ctx, h.logger, w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse webhook event"); respErr != nil {
			if h.logger != nil {
				h.logger.Error(ctx, "failed to send parse error response", "error", respErr.Error())
			}
		}
		return
	}

	// 8. Dispatch to registered handler
	eventTypeStr := string(event.EventType)
	handler, ok := h.handlers[eventTypeStr]
	if !ok {
		// No handler registered for this event type - acknowledge gracefully (202 Accepted)
		if err := respondAccepted(ctx, h.logger, w); err != nil {
			if h.logger != nil {
				h.logger.Error(ctx, "failed to send accepted response for unregistered event type", "error", err.Error(), "eventType", eventTypeStr)
			}
		}
		return
	}

	// Execute handler
	if err := handler(ctx, event); err != nil {
		if respErr := respondError(ctx, h.logger, w, http.StatusInternalServerError, "HANDLER_ERROR", "Failed to process webhook event"); respErr != nil {
			if h.logger != nil {
				h.logger.Error(ctx, "failed to send handler error response", "error", respErr.Error())
			}
		}
		return
	}

	// Mark as processed
	if h.store != nil {
		dedupeKey := dedupeHash(body)
		if markErr := h.store.MarkProcessed(ctx, provider, accountID, dedupeKey); markErr != nil {
			// Log marking error but don't fail - the event was already processed successfully
			// Returning 200 to prevent retries since processing completed
			if h.logger != nil {
				h.logger.Error(ctx, "failed to mark webhook as processed", "error", markErr.Error(), "provider", provider, "accountID", accountID)
			}
		}
	}

	// Success
	if err := respondSuccess(ctx, h.logger, w, "ACK", "Webhook processed successfully"); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "failed to send success response", "error", err.Error())
		}
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
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    errCode,
		"message": message,
	}); err != nil {
		if logger != nil {
			logger.Error(ctx, "failed to encode error response in respondError", "error", err.Error(), "statusCode", statusCode, "errCode", errCode)
		}
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
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	}); err != nil {
		if logger != nil {
			logger.Error(ctx, "failed to encode success response in respondSuccess", "error", err.Error(), "code", code)
		}
		return fmt.Errorf("encode success response: %w", err)
	}
	return nil
}

// respondAccepted sends a JSON accepted response (HTTP 202 Accepted).
// Used when the webhook is valid but no handler is registered for the event type.
// Returns an error if JSON encoding fails (response may be partially written).
func respondAccepted(ctx context.Context, logger ports.Logger, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    "ACCEPTED",
		"message": "Event accepted but not yet processed",
	}); err != nil {
		if logger != nil {
			logger.Error(ctx, "failed to encode accepted response in respondAccepted", "error", err.Error())
		}
		return fmt.Errorf("encode accepted response: %w", err)
	}
	return nil
}
