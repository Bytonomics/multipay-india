# Cashfree Webhook Implementation Notes

## Files Created/Modified

### 1. `providers/cashfree/webhooks.go` (NEW)
Complete implementation of webhook signature verification and event parsing for Cashfree.

#### Key Functions

**`verifySignature(body []byte, headers map[string]string, secret string) error`**
- Verifies HMAC-SHA256 signature of webhook payloads
- Extracts signature from `X-Cashfree-Signature` header (case-insensitive)
- Uses `crypto/hmac` and `crypto/sha256` for cryptographic operations
- Uses `subtle.ConstantTimeCompare` to prevent timing attacks
- Returns `domain.ErrWebhookVerificationFailed` on verification failure
- Validates that both body and secret are non-empty

**`parseEvent(ctx context.Context, body []byte, headers map[string]string) (*domain.WebhookEvent, error)`**
- Unmarshals JSON webhook payload into `cashfreeWebhookPayload` struct
- Maps Cashfree event types to domain `WebhookEventType` enum
- Converts Unix timestamp to `time.Time` (defaults to `time.Now()` if not provided)
- Returns `domain.WebhookEvent` with all required fields populated
- Returns `domain.ErrWebhookEventNotFound` for unsupported event types

**`(a *Adapter) VerifySignature(ctx context.Context, signature string, payload []byte) error`**
- Adapter method implementing `WebhookConsumerProvider` interface
- Delegates to internal `verifySignature` function
- Uses adapter's `ClientSecret` from config

**`(a *Adapter) ParseEvent(ctx context.Context, payload []byte) (*domain.WebhookEvent, error)`**
- Adapter method implementing `WebhookConsumerProvider` interface
- Delegates to internal `parseEvent` function

#### Supported Cashfree Event Types

Mapping defined in `webhookEventMap`:

| Cashfree Event Type | Domain Event Type | Notes |
|-------------------|-------------------|-------|
| `ORDER.PAID` | `EventPaymentCaptured` | Order payment completed |
| `ORDER.EXPIRED` | `EventOrderCreated` | Order expired (treated as order created for now) |
| `PAYMENT.AUTHORIZED` | `EventPaymentAuthorized` | Payment authorized (not captured) |
| `PAYMENT.FAILED` | `EventPaymentFailed` | Payment failed |
| `REFUND.PROCESSED` | `EventRefundProcessed` | Refund completed |
| `REFUND.FAILED` | `EventPaymentFailed` | Refund failed (treated like payment failure) |

**Future Enhancement**: Additional Cashfree event types can be added to `webhookEventMap` as they become relevant.

#### Error Handling

The implementation follows the project's error handling patterns:
- Uses sentinel errors from `domain/errors.go`: `ErrWebhookVerificationFailed`, `ErrWebhookEventNotFound`
- Wraps errors with context using `fmt.Errorf("...: %w", sentinel)`
- All validation errors logged with clear messages
- No panics—all error conditions are explicitly handled

#### Security Features

1. **Timing Attack Prevention**: Uses `subtle.ConstantTimeCompare` for signature comparison
2. **Header Case Insensitivity**: Uses `strings.EqualFold` for header matching
3. **Constant-Time Comparison**: Prevents attackers from determining correct signature byte-by-byte
4. **Empty Input Validation**: Rejects empty payloads and missing secrets

### 2. `providers/cashfree/webhooks_test.go` (NEW)
Comprehensive test suite with 14 test cases and 2 benchmarks.

#### Test Coverage

**Signature Verification Tests**:
- `TestVerifySignature_ValidSignature` - Valid signature accepted
- `TestVerifySignature_InvalidSignature` - Invalid signature rejected
- `TestVerifySignature_MissingHeader` - Missing header rejected
- `TestVerifySignature_CaseInsensitiveHeader` - Case-insensitive header matching
- `TestVerifySignature_EmptyBody` - Empty body rejected

**Event Parsing Tests**:
- `TestParseEvent_OrderPaid` - ORDER.PAID event parsing
- `TestParseEvent_PaymentAuthorized` - PAYMENT.AUTHORIZED event parsing
- `TestParseEvent_RefundProcessed` - REFUND.PROCESSED event parsing
- `TestParseEvent_UnsupportedEventType` - Unsupported event types rejected
- `TestParseEvent_InvalidJSON` - Invalid JSON rejected
- `TestParseEvent_EmptyBody` - Empty body rejected
- `TestParseEvent_DefaultTimestamp` - Missing timestamp defaults to current time

**Adapter Integration Tests**:
- `TestAdapterVerifySignature` - Adapter's VerifySignature method
- `TestAdapterParseEvent` - Adapter's ParseEvent method

**Benchmarks**:
- `BenchmarkVerifySignature` - Signature verification performance
- `BenchmarkParseEvent` - Event parsing performance

All tests follow table-driven test patterns and include edge cases.

### 3. `providers/cashfree/adapter.go` (MODIFIED)
Updated the `VerifySignature` and `ParseEvent` method stubs to call the new webhook implementation.

**Changes**:
- Removed `panic("not implemented")` from `VerifySignature`
- Removed `panic("not implemented")` from `ParseEvent`
- Updated both methods to delegate to internal functions
- Added comments referencing webhooks.go for implementation details

## Implementation Details

### Webhook Payload Structure

The `cashfreeWebhookPayload` struct captures the expected Cashfree webhook JSON structure:
```go
type cashfreeWebhookPayload struct {
    EventID   string                 `json:"event_id"`
    EventType string                 `json:"event_type"`
    CreatedAt int64                  `json:"created_at"`
    Data      map[string]interface{} `json:"data"`
}
```

### Signature Verification Flow

1. Extract `X-Cashfree-Signature` header from request headers
2. Compute HMAC-SHA256 of payload using merchant secret
3. Encode computed hash as hex string
4. Compare with header signature using constant-time comparison
5. Return error if mismatch, nil if valid

### Event Parsing Flow

1. Unmarshal JSON payload into `cashfreeWebhookPayload`
2. Map Cashfree event type string to domain `WebhookEventType` enum
3. Convert Unix timestamp to `time.Time`
4. Construct and return `domain.WebhookEvent`

## Usage Example

```go
// Create adapter
config := &Config{
    ClientID:     "merchant_client_id",
    ClientSecret: "merchant_secret",
    Environment:  domain.EnvironmentProduction,
}
adapter, _ := NewAdapter(config)

// Verify webhook signature (from HTTP headers)
signature := request.Header.Get("X-Cashfree-Signature")
err := adapter.VerifySignature(ctx, signature, body)

// Parse webhook event
event, err := adapter.ParseEvent(ctx, body)

// Access event fields
switch event.EventType {
case domain.EventPaymentCaptured:
    // Handle payment captured
case domain.EventRefundProcessed:
    // Handle refund processed
}
```

## Testing

Run tests:
```bash
cd golang/multipay-adapter/providers/cashfree
go test -v ./... -run Webhook
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./...
```

## Notes for Future Work

1. **Event Type Expansion**: Additional Cashfree event types (PAYMENT.AUTHORIZED, ORDER.EXPIRED, etc.) can be added to `webhookEventMap` as needed
2. **Timestamp Format Handling**: Currently assumes Unix seconds; could be extended to handle other timestamp formats
3. **Event Data Mapping**: The `Data` field contains raw Cashfree event data; could be enhanced with typed extraction helpers
4. **Webhook Endpoint Integration**: This implementation should be called from an HTTP endpoint handler that:
   - Extracts headers and body from request
   - Calls `VerifySignature` first
   - Calls `ParseEvent` to get structured event
   - Processes event based on type
