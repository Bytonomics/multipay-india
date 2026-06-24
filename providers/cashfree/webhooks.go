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

// cfWebhookEnvelope matches Cashfree webhook API version 2023-08-01.
// VERIFIED: top-level key is "type", not "event_type".
type cfWebhookEnvelope struct {
	EventID   string          `json:"event_id"`
	Type      string          `json:"type"`       // "SUBSCRIPTION_STATUS_CHANGED", "ORDER.PAID", etc.
	EventTime string          `json:"event_time"` // RFC3339
	Data      json.RawMessage `json:"data"`
}

// cfSubscriptionStatusChangedData is the data payload for SUBSCRIPTION_STATUS_CHANGED events.
// Status is at data.subscription_details.subscription_status (NOT data.status).
type cfSubscriptionStatusChangedData struct {
	SubscriptionDetails struct {
		SubscriptionID     string `json:"subscription_id"`
		SubscriptionStatus string `json:"subscription_status"` // ACTIVE, ON_HOLD, CUSTOMER_PAUSED, etc.
	} `json:"subscription_details"`
}

// cfSubscriptionEventData is the data payload for most direct subscription event types.
type cfSubscriptionEventData struct {
	SubscriptionDetails struct {
		SubscriptionID string `json:"subscription_id"`
	} `json:"subscription_details"`
}

// cfCardExpiryData is the data payload for SUBSCRIPTION_CARD_EXPIRY_REMINDER.
// It has a different nesting than other subscription events.
type cfCardExpiryData struct {
	SubscriptionStatusWebhook struct {
		SubscriptionDetails struct {
			SubscriptionID string `json:"subscription_id"`
		} `json:"subscription_details"`
	} `json:"subscription_status_webhook"`
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
//   - ctx: Context for the operation (currently unused, reserved for future use)
//   - body: The raw webhook payload bytes (JSON)
//   - headers: The HTTP headers from the webhook request (currently unused, reserved for future use)
//
// Returns:
//   - A *domain.WebhookEvent with Type, Provider, and Data fields populated
//   - An error if JSON unmarshaling fails
func parseEvent(ctx context.Context, body []byte, headers map[string]string) (*domain.WebhookEvent, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("webhook body is empty: %w", domain.ErrWebhookEventNotFound)
	}

	// Unmarshal the JSON payload into the typed envelope.
	var envelope cfWebhookEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook envelope: %w", err)
	}

	// Parse EventTime from RFC3339 string
	var eventTime time.Time
	if envelope.EventTime != "" {
		var err error
		eventTime, err = time.Parse(time.RFC3339, envelope.EventTime)
		if err != nil {
			// Malformed timestamp: degrade gracefully to now (D12: parse errors never abort dispatch)
			eventTime = time.Now()
		}
	} else {
		eventTime = time.Now()
	}

	// Construct base webhook event
	// RawPayload intentionally NOT set here — ServeHTTP sets it to the full verbatim body (D11).
	event := &domain.WebhookEvent{
		Provider:           domain.ProviderCashfree,
		EventTime:          &eventTime,
		DedupeKey:          envelope.EventID,
		RawVendorEventType: envelope.Type, // D11
	}

	// Route by event type
	switch envelope.Type {
	case "SUBSCRIPTION_AUTH_STATUS":
		event.EventType = domain.EventSubAuthenticated
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_PAYMENT_SUCCESS":
		event.EventType = domain.EventSubCharged
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_PAYMENT_FAILED":
		event.EventType = domain.EventSubPaymentFailed
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_REFUND_STATUS":
		event.EventType = domain.EventSubRefund
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_CARD_EXPIRY_REMINDER":
		event.EventType = domain.EventSubCardExpiring
		populateSubscriptionFromCardExpiryData(event, envelope.Data)

	case "SUBSCRIPTION_PAYMENT_CANCELLED":
		event.EventType = domain.EventSubPaymentCancelled
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_PAYMENT_NOTIFICATION_INITIATED":
		event.EventType = domain.EventSubPreDebitNotice
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_PAYMENT_CONTROLLED_NOTIFICATION_STATUS":
		event.EventType = domain.EventSubPreDebitNotice
		populateSubscriptionFromEventData(event, envelope.Data)

	case "SUBSCRIPTION_PAYMENT_CONTROLLED_EXECUTION_STATUS":
		// Route by inner status: SUCCESS → EventSubCharged, FAILED → EventSubPaymentFailed
		type cfControlledExecutionData struct {
			SubscriptionDetails struct {
				SubscriptionID string `json:"subscription_id"`
				Status         string `json:"status"` // SUCCESS or FAILED
			} `json:"subscription_details"`
		}
		var data cfControlledExecutionData
		if err := json.Unmarshal(envelope.Data, &data); err == nil && len(data.SubscriptionDetails.SubscriptionID) > 0 {
			event.Subscription = &domain.Subscription{
				SubscriptionID:         data.SubscriptionDetails.SubscriptionID,
				ProviderSubscriptionID: data.SubscriptionDetails.SubscriptionID,
			}
			event.RawVendorStatus = data.SubscriptionDetails.Status
			// Route by inner status
			switch data.SubscriptionDetails.Status {
			case "SUCCESS":
				event.EventType = domain.EventSubCharged
			case "FAILED":
				event.EventType = domain.EventSubPaymentFailed
			default:
				event.EventType = domain.EventUnknown
			}
		} else {
			event.EventType = domain.EventUnknown
		}

	case "SUBSCRIPTION_STATUS_CHANGED":
		handleSubscriptionStatusChanged(event, envelope.Data)

	case "ORDER.PAID":
		event.EventType = domain.EventPaymentCaptured

	case "ORDER.EXPIRED":
		event.EventType = domain.EventOrderExpired // D16 fix

	case "PAYMENT.AUTHORIZED":
		event.EventType = domain.EventPaymentAuthorized

	case "PAYMENT.FAILED":
		event.EventType = domain.EventPaymentFailed

	case "REFUND.PROCESSED":
		event.EventType = domain.EventRefundProcessed

	case "REFUND.FAILED":
		event.EventType = domain.EventRefundFailed // D16 fix

	default:
		// Unknown event type → EventUnknown, not an error
		event.EventType = domain.EventUnknown
	}

	return event, nil
}

// populateSubscriptionFromEventData unmarshals cfSubscriptionEventData and populates event.Subscription.
func populateSubscriptionFromEventData(event *domain.WebhookEvent, rawData json.RawMessage) {
	var data cfSubscriptionEventData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return
	}

	if data.SubscriptionDetails.SubscriptionID != "" {
		event.Subscription = &domain.Subscription{
			SubscriptionID:         data.SubscriptionDetails.SubscriptionID,
			ProviderSubscriptionID: data.SubscriptionDetails.SubscriptionID,
		}
	}
}

// populateSubscriptionFromCardExpiryData unmarshals cfCardExpiryData and populates event.Subscription.
func populateSubscriptionFromCardExpiryData(event *domain.WebhookEvent, rawData json.RawMessage) {
	var data cfCardExpiryData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return
	}

	if data.SubscriptionStatusWebhook.SubscriptionDetails.SubscriptionID != "" {
		event.Subscription = &domain.Subscription{
			SubscriptionID:         data.SubscriptionStatusWebhook.SubscriptionDetails.SubscriptionID,
			ProviderSubscriptionID: data.SubscriptionStatusWebhook.SubscriptionDetails.SubscriptionID,
		}
	}
}

// handleSubscriptionStatusChanged routes based on subscription status in SUBSCRIPTION_STATUS_CHANGED events.
func handleSubscriptionStatusChanged(event *domain.WebhookEvent, rawData json.RawMessage) {
	var data cfSubscriptionStatusChangedData
	if err := json.Unmarshal(rawData, &data); err != nil {
		event.EventType = domain.EventUnknown
		return
	}

	subscriptionID := data.SubscriptionDetails.SubscriptionID
	status := data.SubscriptionDetails.SubscriptionStatus

	// Populate subscription
	if subscriptionID != "" {
		event.Subscription = &domain.Subscription{
			SubscriptionID:         subscriptionID,
			ProviderSubscriptionID: subscriptionID,
		}
	}

	// Set RawVendorStatus (D11)
	event.RawVendorStatus = status

	// Route by subscription status
	switch status {
	case "ACTIVE":
		event.EventType = domain.EventSubActivated
	case "ON_HOLD":
		event.EventType = domain.EventSubOnHold
	case "CUSTOMER_PAUSED":
		event.EventType = domain.EventSubPaused
	case "CANCELLED", "CUSTOMER_CANCELLED":
		event.EventType = domain.EventSubCancelled
	case "COMPLETED":
		event.EventType = domain.EventSubCompleted
	case "EXPIRED", "LINK_EXPIRED":
		event.EventType = domain.EventSubExpired
	case "BANK_APPROVAL_PENDING":
		event.EventType = domain.EventSubBankApprovalPending
	case "CARD_EXPIRED":
		event.EventType = domain.EventSubCardExpired
	default:
		event.EventType = domain.EventUnknown
	}
}
