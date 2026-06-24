package razorpay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// webhookEventMap contains the Razorpay event type to domain.WebhookEventType mappings.
// Reference: Razorpay webhook events documented in their API reference.
var webhookEventMap = map[string]domain.WebhookEventType{
	"payment.authorized": domain.EventPaymentAuthorized,
	"payment.failed":     domain.EventPaymentFailed,
	"payment.captured":   domain.EventPaymentCaptured,
	"refund.created":     domain.EventRefundCreated,
	"refund.failed":      domain.EventRefundFailed,    // D16 fix — correct mapping
	"refund.processed":   domain.EventRefundProcessed, // G11: general refund event
	// Subscription events
	"subscription.authenticated":  domain.EventSubAuthenticated,
	"subscription.activated":      domain.EventSubActivated,
	"subscription.charged":        domain.EventSubCharged,
	"subscription.payment_failed": domain.EventSubPaymentFailed,
	"subscription.pending":        domain.EventSubOnHold, // pending = subscription at-risk (G1 fix)
	"subscription.halted":         domain.EventSubHalted,
	"subscription.paused":         domain.EventSubPaused,
	"subscription.resumed":        domain.EventSubResumed,
	"subscription.cancelled":      domain.EventSubCancelled,
	"subscription.completed":      domain.EventSubCompleted,
	"subscription.updated":        domain.EventSubUpdated,
}

// razorpayWebhookPayload represents the Razorpay webhook payload structure.
// This is a simplified representation capturing the essential fields.
type razorpayWebhookPayload struct {
	EventID   string         `json:"event_id"`
	Event     string         `json:"event"`
	CreatedAt int64          `json:"created_at"`
	Payload   map[string]any `json:"payload"`
}

// verifySignature verifies the HMAC-SHA256 signature of a Razorpay webhook payload.
// Razorpay sends the signature in the "X-Razorpay-Signature" header.
// The signature is computed as HMAC-SHA256(payload, secret).
//
// Parameters:
//   - body: The raw webhook payload bytes
//   - headers: The HTTP headers from the webhook request (header names should be lowercase or normalized)
//   - secret: The webhook secret (Razorpay webhook secret)
//
// Returns:
//   - nil if the signature is valid
//   - domain.ErrWebhookVerificationFailed if the signature is missing or invalid
func verifySignature(body []byte, headers map[string]string, secret string) error {
	if len(body) == 0 {
		return fmt.Errorf("webhook body is empty: %w", domain.ErrWebhookVerificationFailed)
	}

	if secret == "" {
		return fmt.Errorf("webhook secret is required: %w", domain.ErrWebhookVerificationFailed)
	}

	// Extract the signature from headers.
	// Razorpay uses "X-Razorpay-Signature" header.
	// Normalize header name to lowercase for case-insensitive lookup.
	var signature string
	for key, value := range headers {
		if strings.EqualFold(key, "X-Razorpay-Signature") {
			signature = value
			break
		}
	}

	if signature == "" {
		return fmt.Errorf("X-Razorpay-Signature header is missing: %w", domain.ErrWebhookVerificationFailed)
	}

	// Compute the expected HMAC-SHA256 signature.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
		return fmt.Errorf("signature mismatch: %w", domain.ErrWebhookVerificationFailed)
	}

	return nil
}

// parseEvent parses a Razorpay webhook payload and returns a domain.WebhookEvent.
// The payload is expected to be a JSON object with event metadata and event-specific data.
//
// Parameters:
//   - ctx: Context for the operation (currently unused, reserved for future use)
//   - body: The raw webhook payload bytes (JSON)
//   - headers: The HTTP headers from the webhook request (currently unused, reserved for future use)
//
// Returns:
//   - A *domain.WebhookEvent with Type, Provider, and Data fields populated
//   - An error if JSON unmarshaling fails or the event type is unknown
func parseEvent(ctx context.Context, body []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("webhook body is empty: %w", domain.ErrWebhookEventNotFound)
	}

	// Unmarshal the JSON payload into a temporary structure.
	var payload razorpayWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	// Normalize the event type to lowercase for mapping (Razorpay uses lowercase).
	eventTypeLower := strings.ToLower(payload.Event)

	// Map the Razorpay event type to the canonical domain event type.
	domainEventType, ok := webhookEventMap[eventTypeLower]
	if !ok {
		// Unknown event — route to DefaultHandler, not an error.
		domainEventType = domain.EventUnknown
	}

	// Convert the Unix timestamp (createdAt) to time.Time.
	// Razorpay typically provides timestamps in Unix seconds (not milliseconds).
	var eventTime time.Time
	if payload.CreatedAt > 0 {
		eventTime = time.Unix(payload.CreatedAt, 0)
	} else {
		eventTime = time.Now()
	}

	// Construct base webhook event.
	// RawPayload intentionally NOT set here — ServeHTTP sets it to the full verbatim body (D11).
	event := &domain.WebhookEvent{
		Provider:           domain.ProviderRazorpay,
		EventType:          domainEventType,
		EventTime:          &eventTime,
		DedupeKey:          payload.EventID,
		RawVendorEventType: payload.Event, // verbatim Razorpay event name (D11)
	}

	// Handle subscription webhook events
	if strings.HasPrefix(eventTypeLower, "subscription.") {
		// Extract subscription wrapper from the webhook payload
		if subscriptionWrapper, ok := payload.Payload["subscription"].(map[string]any); ok {
			// Unwrap the entity key (Razorpay nests subscription fields under entity)
			entityData, entityOK := subscriptionWrapper["entity"].(map[string]any)
			if !entityOK || len(entityData) == 0 {
				event.ParseError = "subscription.entity missing or empty in webhook payload"
			} else {
				// Decode subscription from the entity data
				typed, derr := decodeResponse[razorpaySubscriptionResponse](entityData)
				if derr != nil {
					// Decode failure: preserve event type, record parse error (D12: never abort dispatch)
					event.ParseError = derr.Error()
					event.Subscription = nil
				} else {
					event.Subscription = mapSubscriptionFromResponse(typed, entityData)
					// Set RawVendorStatus for subscription events (D11)
					if typed.Status != "" {
						event.RawVendorStatus = typed.Status
					}
				}
			}
		}
	}

	return event, nil
}
