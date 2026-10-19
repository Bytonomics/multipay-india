package ports

import (
	"context"
	"reflect"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
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
// A Logger is mandatory: every constructor that accepts one panics when it is missing, and no fallback
// logger is used. A typed nil (for example a nil *myLogger stored in the interface) also counts as
// missing; use IsNilLogger to test for it.
type Logger interface {
	// Info logs an informational message.
	Info(ctx context.Context, message string, keyvals ...any)

	// Error logs an error message.
	Error(ctx context.Context, message string, keyvals ...any)

	// Debug logs a debug message.
	Debug(ctx context.Context, message string, keyvals ...any)
}

// IsNilLogger reports whether l is missing. It returns true for a nil interface, and also for an
// interface that holds a typed nil: a nil Pointer, Map, Func, Chan or Slice value. Such a value
// passes a plain `l == nil` check, which is why constructors use this function instead.
// Call it only in constructors and config entry points, never on request paths or in method bodies.
func IsNilLogger(l Logger) bool {
	if l == nil {
		return true
	}
	v := reflect.ValueOf(l)
	k := v.Kind()
	nilable := k == reflect.Pointer || k == reflect.Map || k == reflect.Func ||
		k == reflect.Chan || k == reflect.Slice
	return nilable && v.IsNil()
}

// NoopLogger is a no-op Logger. Pass it explicitly when logs should be discarded; a nil Logger is never substituted.
type NoopLogger struct{}

func (l *NoopLogger) Info(ctx context.Context, message string, keyvals ...any)  {}
func (l *NoopLogger) Error(ctx context.Context, message string, keyvals ...any) {}
func (l *NoopLogger) Debug(ctx context.Context, message string, keyvals ...any) {}

func NewNoopLogger() Logger {
	return &NoopLogger{}
}
