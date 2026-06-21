package client

import (
	"errors"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// ClientConfig holds configuration for the MultiPayClient.
// It specifies which payment provider adapters to use, optional hooks for cross-cutting concerns,
// and optional stores/services for webhooks, logging, and time operations.
//
// Adapters must be configured; all other fields are optional (nil = defaults or noop).
type ClientConfig struct {
	// Providers is a list of configured payment provider adapters.
	// At least one provider must be configured and not nil.
	Providers []ports.ProviderAdapter

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
// It checks that at least one provider is configured and that all providers are non-nil.
// Returns an error if validation fails.
func (c *ClientConfig) Validate() error {
	if c == nil {
		return errors.New("config cannot be nil")
	}

	if len(c.Providers) == 0 {
		return errors.New("at least one provider must be configured")
	}

	for i, p := range c.Providers {
		if p == nil {
			return fmt.Errorf("provider at index %d is nil", i)
		}
	}

	return nil
}
