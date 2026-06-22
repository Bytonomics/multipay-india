package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors represent immutable, specific error conditions.
// Each sentinel should be checked with errors.Is() to detect error types.
// Never modify or reuse sentinel values.

var (
	// ErrProviderNotFound is returned when a requested payment provider is not registered
	ErrProviderNotFound = errors.New("provider not found")

	// ErrOrderNotFound is returned when an order with the given ID does not exist
	ErrOrderNotFound = errors.New("order not found")

	// ErrPaymentNotFound is returned when a payment cannot be found for the order
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrRefundNotFound is returned when a refund with the given ID does not exist
	ErrRefundNotFound = errors.New("refund not found")

	// ErrInstrumentNotFound is returned when an instrument/token is not found
	ErrInstrumentNotFound = errors.New("instrument not found")

	// ErrPaymentLinkNotFound is returned when a payment link is not found
	ErrPaymentLinkNotFound = errors.New("payment link not found")

	// ErrInvalidRequest is returned when request validation fails
	ErrInvalidRequest = errors.New("invalid request")

	// ErrUnsupportedCapability is returned when a capability is not supported by the provider
	ErrUnsupportedCapability = errors.New("unsupported capability")

	// ErrProviderError is returned when an upstream provider API error occurs
	ErrProviderError = errors.New("provider error")

	// ErrWebhookVerificationFailed is returned when webhook signature verification fails
	ErrWebhookVerificationFailed = errors.New("webhook verification failed")

	// ErrWebhookEventNotFound is returned when a webhook event cannot be processed
	ErrWebhookEventNotFound = errors.New("webhook event not found")

	// ErrHookPanic is returned when a hook panics during execution
	ErrHookPanic = errors.New("hook panic")
)

// CapabilityError represents an error when a provider does not support a capability.
// It wraps ErrUnsupportedCapability and can be detected with errors.Is().
type CapabilityError struct {
	Provider   Provider
	Capability string
	Message    string
}

// Error implements the error interface.
func (e *CapabilityError) Error() string {
	return fmt.Sprintf("capability error: provider %s does not support %s: %s", e.Provider, e.Capability, e.Message)
}

// Unwrap returns the sentinel error for chain traversal with errors.Is().
func (e *CapabilityError) Unwrap() error {
	return ErrUnsupportedCapability
}

// NewCapabilityError constructs a CapabilityError.
func NewCapabilityError(provider Provider, capability string, msg string) *CapabilityError {
	return &CapabilityError{
		Provider:   provider,
		Capability: capability,
		Message:    msg,
	}
}

// ValidationError represents an error when request validation fails.
// It wraps ErrInvalidRequest and can be detected with errors.Is().
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: field %q: %s", e.Field, e.Message)
	}
	return "validation error: " + e.Message
}

// Unwrap returns the sentinel error for chain traversal with errors.Is().
func (e *ValidationError) Unwrap() error {
	return ErrInvalidRequest
}

// NewValidationError constructs a ValidationError.
func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: msg,
	}
}

// ProviderAPIError represents an error returned by an upstream provider API.
// It wraps ErrProviderError and can be detected with errors.Is().
type ProviderAPIError struct {
	Provider   Provider
	StatusCode int
	ErrorCode  string
	Message    string
	RawBody    []byte
}

// Error implements the error interface.
func (e *ProviderAPIError) Error() string {
	return fmt.Sprintf("provider %s error: %s (status %d, code %q)", e.Provider, e.Message, e.StatusCode, e.ErrorCode)
}

// Unwrap returns the sentinel error for chain traversal with errors.Is().
func (e *ProviderAPIError) Unwrap() error {
	return ErrProviderError
}

// NewProviderAPIError constructs a ProviderAPIError.
func NewProviderAPIError(provider Provider, code int, errCode, msg string) *ProviderAPIError {
	return &ProviderAPIError{
		Provider:   provider,
		StatusCode: code,
		ErrorCode:  errCode,
		Message:    msg,
	}
}

// WebhookError represents an error during webhook processing.
// It wraps ErrWebhookVerificationFailed and can be detected with errors.Is().
type WebhookError struct {
	Reason    string
	Provider  Provider
	AccountID string
}

// Error implements the error interface.
func (e *WebhookError) Error() string {
	return fmt.Sprintf("webhook error: provider %s (account %s): %s", e.Provider, e.AccountID, e.Reason)
}

// Unwrap returns the sentinel error for chain traversal with errors.Is().
func (e *WebhookError) Unwrap() error {
	return ErrWebhookVerificationFailed
}

// HookPanicError represents an error when a hook panics during execution.
// It wraps ErrHookPanic and can be detected with errors.Is().
type HookPanicError struct {
	Phase      string // "Before", "After", "OnError"
	Operation  string // "CreateOrder", etc.
	PanicValue any    // The panic value
	StackTrace string // Full stack trace
}

// Error implements the error interface.
func (e *HookPanicError) Error() string {
	return fmt.Sprintf("hook panic in %s phase for operation %s: %v", e.Phase, e.Operation, e.PanicValue)
}

// Unwrap returns the sentinel error for chain traversal with errors.Is().
func (e *HookPanicError) Unwrap() error {
	return ErrHookPanic
}
