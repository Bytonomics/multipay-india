package ports

import "time"

// Clock provides time-related operations.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// WebhookStore provides durability and idempotency guarantees for webhook processing.
type WebhookStore interface {
	// StoreRawPayload persists the raw webhook payload for audit and recovery.
	// Returns an error if storage fails.
	StoreRawPayload(eventID string, payload []byte) error

	// IsDuplicate checks if a webhook event with the given ID has already been processed.
	// Returns true if the event exists, false otherwise.
	// Returns an error if the check fails.
	IsDuplicate(eventID string) (bool, error)

	// MarkProcessed marks a webhook event as successfully processed.
	// Returns an error if the update fails.
	MarkProcessed(eventID string) error
}

// Logger provides structured logging at multiple severity levels.
type Logger interface {
	// Info logs an informational message.
	Info(message string, keyvals ...any)

	// Error logs an error message.
	Error(message string, keyvals ...any)

	// Debug logs a debug message.
	Debug(message string, keyvals ...any)
}
