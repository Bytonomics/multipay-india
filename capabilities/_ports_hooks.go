// This file is temporarily placed in capabilities/ but should be moved to ports/hooks.go
// It defines Hook interface and HookContext for the ports layer.
package ports

import (
	"context"

	"github.com/Bytonomics/multipay-adapter/domain"
)

// Hook defines the interface for lifecycle hooks around adapter operations.
// Implementations can audit, collect metrics, or enforce cross-cutting concerns.
type Hook interface {
	// Before runs before the adapter executes. Must return nil or an error.
	// If Before returns an error, the operation is short-circuited and OnError is called.
	Before(ctx context.Context, hookCtx *HookContext) error

	// After runs after a successful adapter execution.
	// If After returns an error, it propagates to the caller.
	After(ctx context.Context, hookCtx *HookContext) error

	// OnError runs when the adapter returns an error.
	// err contains the original adapter error.
	// OnError runs even if Before or other hooks failed.
	OnError(ctx context.Context, hookCtx *HookContext, err error) error
}

// HookContext carries contextual information about an operation for hooks.
type HookContext struct {
	// Operation is the name of the operation being executed, e.g. "CreateOrder", "FetchPayment".
	Operation string

	// Provider is the payment provider being used.
	Provider domain.Provider

	// RequestMetadata holds operational metadata from the request.
	// Keys and values are provider-specific and set by the calling code.
	RequestMetadata map[string]interface{}

	// ResponseMetadata holds operational metadata from the response.
	// Populated after successful adapter execution.
	ResponseMetadata map[string]interface{}

	// Error holds the error from the adapter, populated only in OnError.
	Error error
}

// Logger defines the interface for structured logging used by hooks.
type Logger interface {
	// Info logs an informational message.
	Info(msg string)

	// Error logs an error message with optional detail.
	Error(msg, detail string)

	// Warnf logs a formatted warning message.
	Warnf(format string, args ...interface{})
}
