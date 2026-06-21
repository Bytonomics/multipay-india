package hooks

import (
	"context"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// AuditHook logs all payment operations for audit and compliance purposes.
type AuditHook struct {
	logger ports.Logger
}

// NewAuditHook constructs an AuditHook with the provided logger.
func NewAuditHook(logger ports.Logger) *AuditHook {
	return &AuditHook{
		logger: logger,
	}
}

// Before logs the operation before it executes.
func (a *AuditHook) Before(ctx context.Context, hc *ports.HookContext) error {
	if a.logger != nil {
		a.logger.Info(
			"audit: operation starting",
			"provider", hc.Provider.String(),
			"operation", hc.RequestType,
		)
	}
	return nil
}

// After logs the operation after successful completion.
func (a *AuditHook) After(ctx context.Context, hc *ports.HookContext) error {
	if a.logger != nil {
		a.logger.Info(
			"audit: operation completed",
			"provider", hc.Provider.String(),
			"operation", hc.RequestType,
		)
	}
	return nil
}

// OnError logs the operation when it fails.
func (a *AuditHook) OnError(ctx context.Context, hc *ports.HookContext) error {
	if a.logger != nil && hc.Error != nil {
		a.logger.Error(
			"audit: operation failed",
			"provider", hc.Provider.String(),
			"operation", hc.RequestType,
			"error", hc.Error.Error(),
		)
	}
	return nil
}
