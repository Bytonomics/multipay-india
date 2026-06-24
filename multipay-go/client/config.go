package client

import (
	"errors"

	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

// ClientConfig holds configuration for the MultiPayClient.
// It specifies the configured payment provider adapter implementation,
// and optional stores/services for hooks, webhooks, logging, and time operations.
//
// Provider must be configured; all other fields are optional (nil = defaults or noop).
type ClientConfig struct {
	// Provider is the configured payment provider adapter implementation.
	Provider ports.ProviderAdapter

	// Hooks is an optional list of hook middleware.
	// Hooks are executed in FIFO order for Before, After, and OnError phases.
	// If nil or empty, no hooks are executed.
	Hooks []ports.Hook

	// WebhookStore is an optional webhook durability store.
	// If nil, webhooks are not persisted (fire-and-forget semantics).
	WebhookStore ports.WebhookStore

	// Logger is an optional structured logger.
	// If nil, a noop logger is used.
	Logger ports.Logger

	// Clock is an optional time provider.
	// If nil, time.Now is used.
	Clock ports.Clock
}

// Validate validates the ClientConfig.
// It checks that a provider adapter is configured.
// Returns an error if validation fails.
func (c *ClientConfig) Validate() error {
	if c == nil {
		return errors.New("config cannot be nil")
	}

	if c.Provider == nil {
		return errors.New("provider must be configured")
	}

	return nil
}
