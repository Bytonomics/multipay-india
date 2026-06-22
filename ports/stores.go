package ports

import (
	"context"
	"time"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// Clock provides time-related operations.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// RealClock is the production Clock implementation.
type RealClock struct{}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

func NewRealClock() Clock {
	return &RealClock{}
}

// WebhookStore provides durability and idempotency guarantees for webhook processing.
type WebhookStore interface {
	// StoreRawPayload persists the raw webhook payload for audit and recovery.
	// Returns an error if storage fails.
	StoreRawPayload(ctx context.Context, provider domain.Provider, accountID string, payload []byte) error

	// IsDuplicate checks if a webhook event with the given dedupeKey has already been processed.
	// Returns true if the event exists, false otherwise.
	// Returns an error if the check fails.
	IsDuplicate(ctx context.Context, provider domain.Provider, accountID string, dedupeKey string) (bool, error)

	// MarkProcessed marks a webhook event as successfully processed.
	// Returns an error if the update fails.
	MarkProcessed(ctx context.Context, provider domain.Provider, accountID string, dedupeKey string) error
}

// Logger provides structured logging at multiple severity levels.
type Logger interface {
	// Info logs an informational message.
	Info(ctx context.Context, message string, keyvals ...any)

	// Error logs an error message.
	Error(ctx context.Context, message string, keyvals ...any)

	// Debug logs a debug message.
	Debug(ctx context.Context, message string, keyvals ...any)
}

// NoopLogger is a no-op Logger for testing.
type NoopLogger struct{}

func (l *NoopLogger) Info(ctx context.Context, message string, keyvals ...any)  {}
func (l *NoopLogger) Error(ctx context.Context, message string, keyvals ...any) {}
func (l *NoopLogger) Debug(ctx context.Context, message string, keyvals ...any) {}

func NewNoopLogger() Logger {
	return &NoopLogger{}
}
