package routing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// WebhookHandler implements http.Handler for webhook routing and processing.
// It provides the core webhook flow: endpoint matching, deduplication, signature verification,
// event parsing, and handler dispatch.
type WebhookHandler struct {
	matcher  *EndpointMatcher
	store    ports.WebhookStore
	resolver *ports.ProviderRegistry
	handlers map[string]WebhookEventHandler
}

// NewWebhookHandler creates a new WebhookHandler with the given dependencies.
// All dependencies are required (not optional).
func NewWebhookHandler(
	matcher *EndpointMatcher,
	store ports.WebhookStore,
	resolver *ports.ProviderRegistry,
) *WebhookHandler {
	return &WebhookHandler{
		matcher:  matcher,
		store:    store,
		resolver: resolver,
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
	// 1. Match endpoint to extract provider and accountID from path
	provider, accountID, ok := h.matcher.Match(r.URL.Path)
	if !ok {
		respondError(w, http.StatusBadRequest, "INVALID_ENDPOINT", "Invalid webhook endpoint")
		return
	}

	// 2. Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "READ_BODY_FAILED", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// 3. Store raw payload for audit/recovery
	if h.store != nil {
		// Use eventID = provider + ":" + accountID + ":" + dedupeKey for uniqueness
		dedupeKey := dedupeHash(body)
		eventID := fmt.Sprintf("%s:%s:%s", provider, accountID, dedupeKey)
		if err := h.store.StoreRawPayload(eventID, body); err != nil {
			// Log but don't fail - storage is best-effort
			respondError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store webhook payload")
			return
		}

		// 4. Check for duplicate via dedupeKey
		isDuplicate, err := h.store.IsDuplicate(eventID)
		if err != nil {
			// Log but don't fail - dedup check is best-effort
			respondError(w, http.StatusInternalServerError, "DEDUP_ERROR", "Failed to check for duplicates")
			return
		}

		if isDuplicate {
			// Duplicate webhook - return 200 ACK immediately (idempotent)
			respondSuccess(w, "DUPLICATE_ACK", "Duplicate webhook, acknowledged")
			return
		}
	}

	// 5. Resolve adapter for provider
	adapter, err := h.resolver.Resolve(provider)
	if err != nil {
		respondError(w, http.StatusBadRequest, "PROVIDER_NOT_FOUND", fmt.Sprintf("Provider %s not found", provider))
		return
	}

	ctx := r.Context()

	// 6. Verify signature
	signature := extractSignature(r)
	if err := adapter.VerifySignature(ctx, signature, body); err != nil {
		respondError(w, http.StatusBadRequest, "SIGNATURE_INVALID", "Webhook signature verification failed")
		return
	}

	// 7. Parse event
	event, err := adapter.ParseEvent(ctx, body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse webhook event")
		return
	}

	// 8. Dispatch to registered handler
	eventTypeStr := string(event.EventType)
	handler, ok := h.handlers[eventTypeStr]
	if !ok {
		// No handler registered for this event type - acknowledge gracefully (202 Accepted)
		respondAccepted(w)
		return
	}

	// Execute handler
	if err := handler(ctx, event); err != nil {
		respondError(w, http.StatusInternalServerError, "HANDLER_ERROR", "Failed to process webhook event")
		return
	}

	// Mark as processed
	if h.store != nil {
		dedupeKey := dedupeHash(body)
		eventID := fmt.Sprintf("%s:%s:%s", provider, accountID, dedupeKey)
		if err := h.store.MarkProcessed(eventID); err != nil {
			// Log but don't fail - marking as processed is best-effort
			// We've already successfully processed the event, so return success
		}
	}

	// Success
	respondSuccess(w, "ACK", "Webhook processed successfully")
}

// dedupeHash computes a SHA256 hash of the body as a hex string.
// Used to uniquely identify duplicate webhook payloads.
func dedupeHash(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

// extractSignature extracts the webhook signature from the request headers.
// Tries multiple provider-specific signature headers in order:
// - X-Razorpay-Signature (Razorpay)
// - X-Cashfree-Signature (Cashfree)
// Returns empty string if no signature header is found.
func extractSignature(r *http.Request) string {
	// Try Razorpay header
	if sig := r.Header.Get("X-Razorpay-Signature"); sig != "" {
		return sig
	}
	// Try Cashfree header
	if sig := r.Header.Get("X-Cashfree-Signature"); sig != "" {
		return sig
	}
	return ""
}

// respondError sends a JSON error response with the given HTTP status code.
// All error responses have "code" and "message" fields.
func respondError(w http.ResponseWriter, statusCode int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    errCode,
		"message": message,
	})
}

// respondSuccess sends a JSON success response (HTTP 200 OK).
// Response has "code" and "message" fields.
func respondSuccess(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

// respondAccepted sends a JSON accepted response (HTTP 202 Accepted).
// Used when the webhook is valid but no handler is registered for the event type.
func respondAccepted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    "ACCEPTED",
		"message": "Event accepted but not yet processed",
	})
}
