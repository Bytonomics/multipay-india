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
	body := []byte(`{"event_id":"evt_123","event_type":"ORDER.PAID","data":{}}`)

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
	body := []byte(`{"event_id":"evt_123","event_type":"ORDER.PAID","data":{}}`)

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
	body := []byte(`{"event_id":"evt_123","event_type":"ORDER.PAID","data":{}}`)

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
	body := []byte(`{"event_id":"evt_123","event_type":"ORDER.PAID","data":{}}`)

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
	now := time.Now().Unix()
	payload := map[string]interface{}{
		"event_id":   "evt_order_paid_123",
		"event_type": "ORDER.PAID",
		"created_at": now,
		"data": map[string]interface{}{
			"order_id": "order_123",
			"amount":   500.0,
		},
	}

	body, _ := json.Marshal(payload)

	event, err := parseEvent(context.Background(), body, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if event.ID != "evt_order_paid_123" {
		t.Errorf("expected event_id 'evt_order_paid_123', got %q", event.ID)
	}

	if event.EventType != domain.EventPaymentCaptured {
		t.Errorf("expected EventType %v, got %v", domain.EventPaymentCaptured, event.EventType)
	}

	if event.Provider != domain.ProviderCashfree.String() {
		t.Errorf("expected provider 'cashfree', got %q", event.Provider)
	}

	if event.Timestamp.Unix() != now {
		t.Errorf("expected timestamp %d, got %d", now, event.Timestamp.Unix())
	}
}

// TestParseEvent_PaymentAuthorized verifies parsing of PAYMENT.AUTHORIZED event.
func TestParseEvent_PaymentAuthorized(t *testing.T) {
	now := time.Now().Unix()
	payload := map[string]interface{}{
		"event_id":   "evt_payment_auth_123",
		"event_type": "PAYMENT.AUTHORIZED",
		"created_at": now,
		"data": map[string]interface{}{
			"payment_id": "payment_123",
		},
	}

	body, _ := json.Marshal(payload)

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
	now := time.Now().Unix()
	payload := map[string]interface{}{
		"event_id":   "evt_refund_123",
		"event_type": "REFUND.PROCESSED",
		"created_at": now,
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

// TestParseEvent_UnsupportedEventType verifies that unsupported event types are rejected.
func TestParseEvent_UnsupportedEventType(t *testing.T) {
	payload := map[string]interface{}{
		"event_id":   "evt_unknown_123",
		"event_type": "UNKNOWN.EVENT",
		"created_at": time.Now().Unix(),
		"data":       map[string]interface{}{},
	}

	body, _ := json.Marshal(payload)

	_, err := parseEvent(context.Background(), body, nil)
	if err == nil {
		t.Fatal("expected error for unsupported event type, got nil")
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
		"event_type": "ORDER.PAID",
		"created_at": 0,
		"data":       map[string]interface{}{},
	}

	body, _ := json.Marshal(payload)

	before := time.Now()
	event, err := parseEvent(context.Background(), body, nil)
	after := time.Now()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Verify that timestamp is within the expected range (before to after).
	if event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Errorf("expected timestamp between %v and %v, got %v", before, after, event.Timestamp)
	}
}

// TestAdapterVerifySignature verifies the adapter's VerifySignature method.
func TestAdapterVerifySignature(t *testing.T) {
	config := &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	secret := config.ClientSecret
	body := []byte(`{"event_id":"evt_123","event_type":"ORDER.PAID","data":{}}`)

	// Compute the expected signature.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Verify the signature using the adapter method.
	err = adapter.VerifySignature(context.Background(), signature, body)
	if err != nil {
		t.Fatalf("expected nil error from adapter.VerifySignature, got %v", err)
	}
}

// TestAdapterParseEvent verifies the adapter's ParseEvent method.
func TestAdapterParseEvent(t *testing.T) {
	config := &Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		Environment:  domain.EnvironmentSandbox,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	now := time.Now().Unix()
	payload := map[string]interface{}{
		"event_id":   "evt_123",
		"event_type": "ORDER.PAID",
		"created_at": now,
		"data": map[string]interface{}{
			"order_id": "order_123",
		},
	}

	body, _ := json.Marshal(payload)

	event, err := adapter.ParseEvent(context.Background(), body)
	if err != nil {
		t.Fatalf("expected nil error from adapter.ParseEvent, got %v", err)
	}

	if event.Provider != domain.ProviderCashfree.String() {
		t.Errorf("expected provider 'cashfree', got %q", event.Provider)
	}

	if event.EventType != domain.EventPaymentCaptured {
		t.Errorf("expected EventType %v, got %v", domain.EventPaymentCaptured, event.EventType)
	}
}

// BenchmarkVerifySignature benchmarks the signature verification.
func BenchmarkVerifySignature(b *testing.B) {
	secret := "test_secret_key"
	body := []byte(`{"event_id":"evt_123","event_type":"ORDER.PAID","data":{}}`)

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
	payload := map[string]interface{}{
		"event_id":   "evt_123",
		"event_type": "ORDER.PAID",
		"created_at": time.Now().Unix(),
		"data": map[string]interface{}{
			"order_id": "order_123",
			"amount":   500.0,
		},
	}

	body, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseEvent(context.Background(), body, nil)
	}
}
