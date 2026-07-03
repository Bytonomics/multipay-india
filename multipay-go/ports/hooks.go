package ports

import (
	"context"
	"time"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
)

// HookContext contains metadata for a hook execution.
type HookContext struct {
	// Provider is the payment provider for this hook.
	Provider domain.Provider

	// RequestType is the type of request being made (e.g., "CreateOrder", "CreatePayment").
	RequestType string

	// RequestData contains the raw request data passed to the hook.
	RequestData any

	// ResponseData contains the raw response data (for After/OnError hooks).
	ResponseData any

	// Error is the error that occurred (only set for OnError hooks).
	Error error

	// StartTime is when hook execution began.
	StartTime time.Time

	// Metadata is for hooks to pass arbitrary data down the pipeline.
	Metadata map[string]any
}

// Hook defines lifecycle hooks for payment operations.
// Hooks allow for cross-cutting concerns like logging, validation, retry logic,
// and instrumentation without modifying core provider logic.
type Hook interface {
	// Before is called before a payment operation is executed.
	// It can inspect and potentially modify the request context.
	// The returned context is used for the subsequent operation.
	// Returning an error from Before will prevent the operation from executing.
	Before(ctx context.Context, hc *HookContext) (context.Context, error)

	// After is called after a payment operation completes successfully.
	// It can inspect the response and perform post-processing.
	After(ctx context.Context, hc *HookContext) error

	// OnError is called when a payment operation fails.
	// It has access to the error and can perform recovery or logging.
	OnError(ctx context.Context, hc *HookContext) error
}
