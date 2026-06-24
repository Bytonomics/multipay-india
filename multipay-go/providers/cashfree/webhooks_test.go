package cashfree

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// TestVerifySignature_ValidSignature verifies that a valid HMAC-SHA256 signature is accepted.
func TestVerifySignature_ValidSignature(t *testing.T) {
	secret := "test_secret_key"
	body := []byte(`{"event_id":"evt_123","type":"ORDER.PAID","data":{}}`)

	// Compute the expected signature.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Verify the signature.
	headers := map[string]string{
		"X-Cashfree-Signature": signature,
	}

	err := verifySignature(body, headers, secret)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// TestVerifySignature_InvalidSignature verifies that an invalid signature is rejected.
func TestVerifySignature_InvalidSignature(t *testing.T) {
	secret := "test_secret_key"
	body := []byte(`{"event_id":"evt_123","type":"ORDER.PAID","data":{}}`)

	// Use a wrong signature.
	headers := map[string]string{
		"X-Cashfree-Signature": "invalid_signature_hash",
	}

	err := verifySignature(body, headers, secret)
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
}

// TestVerifySignature_MissingHeader verifies that a missing signature header is rejected.
func TestVerifySignature_MissingHeader(t *testing.T) {
	secret := "test_secret_key"
	body := []byte(`{"event_id":"evt_123","type":"ORDER.PAID","data":{}}`)

	// No signature header.
	headers := map[string]string{}

	err := verifySignature(body, headers, secret)
	if err == nil {
		t.Fatal("expected error for missing signature header, got nil")
	}
}

// TestVerifySignature_CaseInsensitiveHeader verifies that header names are case-insensitive.
func TestVerifySignature_CaseInsensitiveHeader(t *testing.T) {
	secret := "test_secret_key"
	body := []byte(`{"event_id":"evt_123","type":"ORDER.PAID","data":{}}`)

	// Compute the expected signature.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Use lowercase header name.
	headers := map[string]string{
		"x-cashfree-signature": signature,
	}

	err := verifySignature(body, headers, secret)
	if err != nil {
		t.Fatalf("expected nil error with case-insensitive header, got %v", err)
	}
}

// TestVerifySignature_EmptyBody verifies that empty body is rejected.
func TestVerifySignature_EmptyBody(t *testing.T) {
	secret := "test_secret_key"
	body := []byte{}

	headers := map[string]string{
		"X-Cashfree-Signature": "any_signature",
	}

	err := verifySignature(body, headers, secret)
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
}

// TestParseEvent_OrderPaid verifies parsing of ORDER.PAID event.
func TestParseEvent_OrderPaid(t *testing.T) {
	now := time.Now()
	eventTimeStr := now.Format(time.RFC3339)
	payload := map[string]interface{}{
		"event_id":   "evt_order_paid_123",
		"type":       "ORDER.PAID",
		"event_time": eventTimeStr,
		"data": map[string]interface{}{
			"order_id": "order_123",
			"amount":   500.0,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if event.Provider != domain.ProviderCashfree {
		t.Errorf("expected provider %v, got %v", domain.ProviderCashfree, event.Provider)
	}

	if event.EventType != domain.EventPaymentCaptured {
		t.Errorf("expected EventType %v, got %v", domain.EventPaymentCaptured, event.EventType)
	}

	if event.EventTime == nil || event.EventTime.Unix() != now.Unix() {
		if event.EventTime == nil {
			t.Error("expected non-nil EventTime")
		} else {
			t.Errorf("expected timestamp %d, got %d", now.Unix(), event.EventTime.Unix())
		}
	}
}

// TestParseEvent_PaymentAuthorized verifies parsing of PAYMENT.AUTHORIZED event.
func TestParseEvent_PaymentAuthorized(t *testing.T) {
	now := time.Now()
	eventTimeStr := now.Format(time.RFC3339)
	payload := map[string]interface{}{
		"event_id":   "evt_payment_auth_123",
		"type":       "PAYMENT.AUTHORIZED",
		"event_time": eventTimeStr,
		"data": map[string]interface{}{
			"payment_id": "payment_123",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if event.EventType != domain.EventPaymentAuthorized {
		t.Errorf("expected EventType %v, got %v", domain.EventPaymentAuthorized, event.EventType)
	}
}

// TestParseEvent_RefundProcessed verifies parsing of REFUND.PROCESSED event.
func TestParseEvent_RefundProcessed(t *testing.T) {
	now := time.Now()
	eventTimeStr := now.Format(time.RFC3339)
	payload := map[string]interface{}{
		"event_id":   "evt_refund_123",
		"type":       "REFUND.PROCESSED",
		"event_time": eventTimeStr,
		"data": map[string]interface{}{
			"refund_id": "refund_123",
			"amount":    100.0,
		},
	}

	body, _ := json.Marshal(payload)

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if event.EventType != domain.EventRefundProcessed {
		t.Errorf("expected EventType %v, got %v", domain.EventRefundProcessed, event.EventType)
	}
}

// TestParseEvent_UnsupportedEventType verifies that unsupported event types gracefully map to EventUnknown.
func TestParseEvent_UnsupportedEventType(t *testing.T) {
	eventTimeStr := time.Now().Format(time.RFC3339)
	payload := map[string]interface{}{
		"event_id":   "evt_unknown_123",
		"type":       "UNKNOWN.EVENT",
		"event_time": eventTimeStr,
		"data":       map[string]interface{}{},
	}

	body, _ := json.Marshal(payload) //nolint:not-an-error

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error for unknown event type, got %v", err)
	}
	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.EventType != domain.EventUnknown {
		t.Errorf("expected EventUnknown for unsupported event type, got %v", event.EventType)
	}
}

// TestParseEvent_InvalidJSON verifies that invalid JSON is rejected.
func TestParseEvent_InvalidJSON(t *testing.T) {
	body := []byte(`{"event_id": "evt_123", invalid json}`)

	_, err := parseEvent(context.Background(), body, nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestParseEvent_EmptyBody verifies that empty body is rejected.
func TestParseEvent_EmptyBody(t *testing.T) {
	body := []byte{}

	_, err := parseEvent(context.Background(), body, nil)
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
}

// TestParseEvent_DefaultTimestamp verifies that missing timestamp defaults to current time.
func TestParseEvent_DefaultTimestamp(t *testing.T) {
	payload := map[string]interface{}{
		"event_id":   "evt_123",
		"type":       "ORDER.PAID",
		"event_time": "",
		"data":       map[string]interface{}{},
	}

	body, _ := json.Marshal(payload) //nolint:not-an-error

	before := time.Now()
	event, err := parseEvent(context.Background(), body, nil)
	after := time.Now()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Verify that timestamp is within the expected range (before to after).
	if event.EventTime.Before(before) || event.EventTime.After(after) {
		t.Errorf("expected timestamp between %v and %v, got %v", before, after, event.EventTime)
	}
}

// TestAdapterVerifySignature verifies the adapter's VerifySignature method.
func TestAdapterVerifySignature(t *testing.T) {
	webhookSecret := "test_webhook_secret"
	config := &Config{
		ClientID:      "test_client_id",
		ClientSecret:  "test_client_secret",
		WebhookSecret: webhookSecret,
		Environment:   domain.EnvironmentSandbox,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	body := []byte(`{"event_id":"evt_123","type":"ORDER.PAID","data":{}}`)

	// Compute the expected signature using the webhook secret.
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Verify the signature using the adapter method.
	headers := map[string]string{
		"X-Cashfree-Signature": signature,
	}
	err = adapter.VerifySignature(context.Background(), body, headers)
	if err != nil {
		t.Fatalf("expected nil error from adapter.VerifySignature, got %v", err)
	}
}

// TestAdapterParseEvent verifies the adapter's ParseEvent method.
func TestAdapterParseEvent(t *testing.T) {
	config := &Config{
		ClientID:      "test_client_id",
		ClientSecret:  "test_client_secret",
		WebhookSecret: "test_webhook_secret",
		Environment:   domain.EnvironmentSandbox,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	now := time.Now()
	eventTimeStr := now.Format(time.RFC3339)
	payload := map[string]interface{}{
		"event_id":   "evt_123",
		"type":       "ORDER.PAID",
		"event_time": eventTimeStr,
		"data": map[string]interface{}{
			"order_id": "order_123",
		},
	}

	body, _ := json.Marshal(payload) //nolint:not-an-error

	event, err := adapter.ParseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error from adapter.ParseEvent, got %v", err)
	}

	if event.Provider != domain.ProviderCashfree {
		t.Errorf("expected provider %v, got %v", domain.ProviderCashfree, event.Provider)
	}

	if event.EventType != domain.EventPaymentCaptured {
		t.Errorf("expected EventType %v, got %v", domain.EventPaymentCaptured, event.EventType)
	}
}

// BenchmarkVerifySignature benchmarks the signature verification.
func BenchmarkVerifySignature(b *testing.B) {
	secret := "test_secret_key"
	body := []byte(`{"event_id":"evt_123","type":"ORDER.PAID","data":{}}`)

	// Compute the expected signature once.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	headers := map[string]string{
		"X-Cashfree-Signature": signature,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifySignature(body, headers, secret)
	}
}

// BenchmarkParseEvent benchmarks the event parsing.
func BenchmarkParseEvent(b *testing.B) {
	eventTimeStr := time.Now().Format(time.RFC3339)
	payload := map[string]interface{}{
		"event_id":   "evt_123",
		"type":       "ORDER.PAID",
		"event_time": eventTimeStr,
		"data": map[string]interface{}{
			"order_id": "order_123",
			"amount":   500.0,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		b.Fatalf("failed to marshal payload: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseEvent(context.Background(), body, nil) //nolint:not-an-error
	}
}

// TestCashfreeParseEvent_DirectSubscriptionEvents tests all direct subscription event types
// (not SUBSCRIPTION_STATUS_CHANGED).
func TestCashfreeParseEvent_DirectSubscriptionEvents(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		expectedType domain.WebhookEventType
	}{
		{"SUBSCRIPTION_AUTH_STATUS", "SUBSCRIPTION_AUTH_STATUS", domain.EventSubAuthenticated},
		{"SUBSCRIPTION_PAYMENT_SUCCESS", "SUBSCRIPTION_PAYMENT_SUCCESS", domain.EventSubCharged},
		{"SUBSCRIPTION_PAYMENT_FAILED", "SUBSCRIPTION_PAYMENT_FAILED", domain.EventSubPaymentFailed},
		{"SUBSCRIPTION_REFUND_STATUS", "SUBSCRIPTION_REFUND_STATUS", domain.EventSubRefund},
		{"SUBSCRIPTION_PAYMENT_CANCELLED", "SUBSCRIPTION_PAYMENT_CANCELLED", domain.EventSubPaymentCancelled},
		{"SUBSCRIPTION_PAYMENT_NOTIFICATION_INITIATED", "SUBSCRIPTION_PAYMENT_NOTIFICATION_INITIATED", domain.EventSubPreDebitNotice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"type":       tt.eventType,
				"event_id":   "evt_1",
				"event_time": "2024-01-01T00:00:00Z",
				"data": map[string]interface{}{
					"subscription_details": map[string]interface{}{
						"subscription_id": "sub_1",
					},
				},
			}
			body, _ := json.Marshal(payload) //nolint:not-an-error

			event, err := parseEvent(context.Background(), body, nil)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", tt.eventType, err)
			}
			if event == nil {
				t.Fatal("expected non-nil event")
			}
			if event.EventType != tt.expectedType {
				t.Errorf("expected %v, got %v", tt.expectedType, event.EventType)
			}
			if event.Subscription == nil {
				t.Error("expected non-nil Subscription")
			} else if event.Subscription.SubscriptionID != "sub_1" {
				t.Errorf("expected SubscriptionID 'sub_1', got %q", event.Subscription.SubscriptionID)
			}
		})
	}
}

// TestCashfreeParseEvent_CardExpiry tests SUBSCRIPTION_CARD_EXPIRY_REMINDER with its different
// payload nesting (subscription_status_webhook instead of direct data).
func TestCashfreeParseEvent_CardExpiry(t *testing.T) {
	payload := map[string]interface{}{
		"type":       "SUBSCRIPTION_CARD_EXPIRY_REMINDER",
		"event_id":   "evt_1",
		"event_time": "2024-01-01T00:00:00Z",
		"data": map[string]interface{}{
			"subscription_status_webhook": map[string]interface{}{
				"subscription_details": map[string]interface{}{
					"subscription_id": "sub_2",
				},
			},
		},
	}
	body, _ := json.Marshal(payload) //nolint:not-an-error

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("failed to parse SUBSCRIPTION_CARD_EXPIRY_REMINDER: %v", err)
	}
	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.EventType != domain.EventSubCardExpiring {
		t.Errorf("expected %v, got %v", domain.EventSubCardExpiring, event.EventType)
	}
	if event.Subscription == nil {
		t.Fatal("expected non-nil Subscription")
	}
	if event.Subscription.SubscriptionID != "sub_2" {
		t.Errorf("expected SubscriptionID 'sub_2', got %q", event.Subscription.SubscriptionID)
	}
}

// TestCashfreeParseEvent_StatusChanged_AllStatuses tests SUBSCRIPTION_STATUS_CHANGED routing
// by subscription_details.subscription_status for all possible status values.
func TestCashfreeParseEvent_StatusChanged_AllStatuses(t *testing.T) {
	tests := []struct {
		name         string
		cfStatus     string
		expectedType domain.WebhookEventType
	}{
		{"ACTIVE", "ACTIVE", domain.EventSubActivated},
		{"ON_HOLD", "ON_HOLD", domain.EventSubOnHold},
		{"CUSTOMER_PAUSED", "CUSTOMER_PAUSED", domain.EventSubPaused},
		{"CANCELLED", "CANCELLED", domain.EventSubCancelled},
		{"CUSTOMER_CANCELLED", "CUSTOMER_CANCELLED", domain.EventSubCancelled},
		{"COMPLETED", "COMPLETED", domain.EventSubCompleted},
		{"EXPIRED", "EXPIRED", domain.EventSubExpired},
		{"LINK_EXPIRED", "LINK_EXPIRED", domain.EventSubExpired},
		{"BANK_APPROVAL_PENDING", "BANK_APPROVAL_PENDING", domain.EventSubBankApprovalPending},
		{"CARD_EXPIRED", "CARD_EXPIRED", domain.EventSubCardExpired},
		{"UNKNOWN_STATUS", "UNKNOWN_STATUS", domain.EventUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"type":       "SUBSCRIPTION_STATUS_CHANGED",
				"event_id":   "evt_1",
				"event_time": "2024-01-01T00:00:00Z",
				"data": map[string]interface{}{
					"subscription_details": map[string]interface{}{
						"subscription_id":     "sub_1",
						"subscription_status": tt.cfStatus,
					},
				},
			}
			body, _ := json.Marshal(payload) //nolint:not-an-error

			event, err := parseEvent(context.Background(), body, nil)
			if err != nil {
				t.Fatalf("failed to parse status %s: %v", tt.cfStatus, err)
			}
			if event == nil {
				t.Fatal("expected non-nil event")
			}
			if event.EventType != tt.expectedType {
				t.Errorf("expected %v for status %q, got %v", tt.expectedType, tt.cfStatus, event.EventType)
			}
			if tt.expectedType != domain.EventUnknown {
				if event.RawVendorEventType != "SUBSCRIPTION_STATUS_CHANGED" {
					t.Errorf("expected RawVendorEventType 'SUBSCRIPTION_STATUS_CHANGED', got %q", event.RawVendorEventType)
				}
			}
			if tt.cfStatus == "ACTIVE" {
				if event.RawVendorStatus != "ACTIVE" {
					t.Errorf("expected RawVendorStatus 'ACTIVE', got %q", event.RawVendorStatus)
				}
			}
		})
	}
}

// TestCashfreeParseEvent_ActiveNoResumeLogic ensures ACTIVE status always maps to
// EventSubActivated regardless of context (no previous_status logic).
func TestCashfreeParseEvent_ActiveNoResumeLogic(t *testing.T) {
	payload := map[string]interface{}{
		"type":       "SUBSCRIPTION_STATUS_CHANGED",
		"event_id":   "evt_1",
		"event_time": "2024-01-01T00:00:00Z",
		"data": map[string]interface{}{
			"subscription_details": map[string]interface{}{
				"subscription_status": "ACTIVE",
			},
		},
	}
	body, _ := json.Marshal(payload) //nolint:not-an-error

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("failed to parse ACTIVE status: %v", err)
	}
	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.EventType != domain.EventSubActivated {
		t.Errorf("expected %v for ACTIVE, got %v", domain.EventSubActivated, event.EventType)
	}
}

// TestCashfreeParseEvent_DirectOrderAndRefundEvents tests order, payment, and refund
// event type routing.
func TestCashfreeParseEvent_DirectOrderAndRefundEvents(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		expectedType domain.WebhookEventType
	}{
		{"ORDER.PAID", "ORDER.PAID", domain.EventPaymentCaptured},
		{"ORDER.EXPIRED", "ORDER.EXPIRED", domain.EventOrderExpired},
		{"PAYMENT.AUTHORIZED", "PAYMENT.AUTHORIZED", domain.EventPaymentAuthorized},
		{"PAYMENT.FAILED", "PAYMENT.FAILED", domain.EventPaymentFailed},
		{"REFUND.PROCESSED", "REFUND.PROCESSED", domain.EventRefundProcessed},
		{"REFUND.FAILED", "REFUND.FAILED", domain.EventRefundFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"type":     tt.eventType,
				"event_id": "e1",
				"data":     map[string]interface{}{},
			}
			body, _ := json.Marshal(payload) //nolint:not-an-error

			event, err := parseEvent(context.Background(), body, nil)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", tt.eventType, err)
			}
			if event == nil {
				t.Fatal("expected non-nil event")
			}
			if event.EventType != tt.expectedType {
				t.Errorf("expected %v for %q, got %v", tt.expectedType, tt.eventType, event.EventType)
			}
		})
	}
}

// TestCashfreeParseEvent_UnknownType tests that an unknown event type returns EventUnknown
// without error.
func TestCashfreeParseEvent_UnknownType(t *testing.T) {
	payload := map[string]interface{}{
		"type":     "SOME_FUTURE_EVENT_TYPE",
		"event_id": "e1",
		"data":     map[string]interface{}{},
	}
	body, _ := json.Marshal(payload) //nolint:not-an-error

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error for unknown type, got %v", err)
	}
	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.EventType != domain.EventUnknown {
		t.Errorf("expected %v for unknown type, got %v", domain.EventUnknown, event.EventType)
	}
}

// TestCashfreeParseEvent_D11Fields verifies that parseEvent sets the parse-level D11 fields
// (RawVendorEventType, RawVendorStatus). RawPayload is intentionally NOT set by parseEvent — the
// HTTP handler (ServeHTTP) populates it with the full verbatim body; that is covered by the
// routing-level TestWebhookHandler_RawPayloadAndHeadersPopulated.
func TestCashfreeParseEvent_D11Fields(t *testing.T) {
	payload := map[string]interface{}{
		"type":       "SUBSCRIPTION_STATUS_CHANGED",
		"event_id":   "evt_1",
		"event_time": "2024-01-01T00:00:00Z",
		"data": map[string]interface{}{
			"subscription_details": map[string]interface{}{
				"subscription_id":     "sub_1",
				"subscription_status": "ACTIVE",
			},
		},
	}
	body, _ := json.Marshal(payload) //nolint:not-an-error

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}
	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.RawVendorEventType != "SUBSCRIPTION_STATUS_CHANGED" {
		t.Errorf("expected RawVendorEventType 'SUBSCRIPTION_STATUS_CHANGED', got %q", event.RawVendorEventType)
	}
	if event.RawVendorStatus != "ACTIVE" {
		t.Errorf("expected RawVendorStatus 'ACTIVE', got %q", event.RawVendorStatus)
	}
}
