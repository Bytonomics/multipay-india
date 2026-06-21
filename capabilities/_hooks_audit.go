// This file is temporarily placed in capabilities/ but should be moved to hooks/audit.go
package hooks

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// AuditHook logs operation details before, after, and on error for audit trails.
// It implements the ports.Hook interface.
type AuditHook struct {
	logger ports.Logger
}

// NewAuditHook constructs an AuditHook with a logger.
// The logger must not be nil.
func NewAuditHook(logger ports.Logger) *AuditHook {
	return &AuditHook{
		logger: logger,
	}
}

// Before logs the operation name, provider, and request metadata.
func (h *AuditHook) Before(ctx context.Context, hookCtx *ports.HookContext) error {
	if h.logger == nil {
		return nil
	}

	msg := fmt.Sprintf("operation %q starting on provider %q", hookCtx.Operation, hookCtx.Provider)
	h.logger.Info(msg)

	// Log request metadata if present.
	if len(hookCtx.RequestMetadata) > 0 {
		detail := fmt.Sprintf("request metadata: %+v", hookCtx.RequestMetadata)
		h.logger.Warnf("audit: %s", detail)
	}

	return nil
}

// After logs successful operation completion and response metadata size.
func (h *AuditHook) After(ctx context.Context, hookCtx *ports.HookContext) error {
	if h.logger == nil {
		return nil
	}

	msg := fmt.Sprintf("operation %q completed successfully", hookCtx.Operation)
	h.logger.Info(msg)

	// Log response metadata size if present.
	if len(hookCtx.ResponseMetadata) > 0 {
		detail := fmt.Sprintf("response metadata has %d fields", len(hookCtx.ResponseMetadata))
		h.logger.Warnf("audit: %s", detail)
	}

	return nil
}

// OnError logs the error and its details for audit trails.
func (h *AuditHook) OnError(ctx context.Context, hookCtx *ports.HookContext, err error) error {
	if h.logger == nil {
		return nil
	}

	msg := fmt.Sprintf("operation %q failed with error", hookCtx.Operation)
	detail := fmt.Sprintf("error: %v", err)
	h.logger.Error(msg, detail)

	return nil
}

// Compile-time check: AuditHook implements ports.Hook.
var _ ports.Hook = (*AuditHook)(nil)
