package cashfree

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

	"github.com/Bytonomics/multipay-adapter/domain"
)

// webhookEventMap contains the Cashfree event type to domain.WebhookEventType mappings.
// Reference: Cashfree webhook events documented in their API reference.
var webhookEventMap = map[string]domain.WebhookEventType{
	"ORDER.PAID":         domain.EventPaymentCaptured,
	"ORDER.EXPIRED":      domain.EventOrderCreated,
	"PAYMENT.AUTHORIZED": domain.EventPaymentAuthorized,
	"PAYMENT.FAILED":     domain.EventPaymentFailed,
	"REFUND.PROCESSED":   domain.EventRefundProcessed,
	"REFUND.FAILED":      domain.EventPaymentFailed, // Treat refund failure similar to payment failure
}

// cashfreeWebhookPayload represents the Cashfree webhook payload structure.
// This is a simplified representation capturing the essential fields.
type cashfreeWebhookPayload struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	CreatedAt int64                  `json:"created_at"`
	Data      map[string]interface{} `json:"data"`
}

// verifySignature verifies the HMAC-SHA256 signature of a Cashfree webhook payload.
// Cashfree sends the signature in the "X-Cashfree-Signature" header.
// The signature is computed as HMAC-SHA256(payload, secret).
//
// Parameters:
//   - body: The raw webhook payload bytes
//   - headers: The HTTP headers from the webhook request (header names should be lowercase or normalized)
//   - secret: The webhook secret (Cashfree merchant secret)
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
	// Cashfree uses "X-Cashfree-Signature" header.
	// Normalize header name to lowercase for case-insensitive lookup.
	var signature string
	for key, value := range headers {
		if strings.EqualFold(key, "X-Cashfree-Signature") {
			signature = value
			break
		}
	}

	if signature == "" {
		return fmt.Errorf("X-Cashfree-Signature header is missing: %w", domain.ErrWebhookVerificationFailed)
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

// parseEvent parses a Cashfree webhook payload and returns a domain.WebhookEvent.
// The payload is expected to be a JSON object with event metadata and event-specific data.
//
// Parameters:
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
	var payload cashfreeWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	// Normalize the event type to lowercase for mapping (Cashfree uses uppercase).
	eventTypeUpper := strings.ToUpper(payload.EventType)

	// Map the Cashfree event type to the canonical domain event type.
	domainEventType, ok := webhookEventMap[eventTypeUpper]
	if !ok {
		// Return an error if the event type is not supported.
		// This is not a verification failure, but an unrecognized event.
		return nil, fmt.Errorf("unsupported Cashfree event type: %q: %w", payload.EventType, domain.ErrWebhookEventNotFound)
	}

	// Convert the Unix timestamp (createdAt) to time.Time.
	// Cashfree typically provides timestamps in Unix seconds (not milliseconds).
	var eventTime time.Time
	if payload.CreatedAt > 0 {
		eventTime = time.Unix(payload.CreatedAt, 0)
	} else {
		eventTime = time.Now()
	}

	// Construct and return the domain webhook event.
	event := &domain.WebhookEvent{
		ID:        payload.EventID,
		EventType: domainEventType,
		Provider:  domain.ProviderCashfree.String(),
		Data:      payload.Data,
		Timestamp: eventTime,
	}

	return event, nil
}

// VerifySignature verifies the authenticity of a webhook request from Cashfree.
// This is the adapter method that wraps the verifySignature helper function.
//
// Parameters:
//   - ctx: Context for the operation (currently unused, reserved for future use)
//   - signature: The signature value (extracted from the X-Cashfree-Signature header by the caller)
//   - payload: The raw webhook payload bytes
//
// Note: The signature parameter here is the raw header value. The actual verification
// is delegated to verifySignature which expects headers as a map. For compatibility
// with the WebhookConsumerProvider interface, this method reconstructs the headers map.
