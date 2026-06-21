// This file is temporarily placed in capabilities/ but should be moved to hooks/pipeline.go
package hooks

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// Pipeline manages execution of Hook instances in FIFO order.
// It handles Before, After, and OnError hook phases with proper short-circuiting and panic recovery.
type Pipeline struct {
	hooks []ports.Hook
}

// NewPipeline constructs a Pipeline with an ordered slice of hooks.
// Hooks will execute in the order provided.
func NewPipeline(hooks ...ports.Hook) *Pipeline {
	return &Pipeline{
		hooks: hooks,
	}
}

// ExecuteBefore runs all Before hooks in FIFO order.
// Returns immediately (short-circuits) if any hook returns an error.
// If a hook panics, the panic is recovered, logged, and converted to an error.
func (p *Pipeline) ExecuteBefore(ctx context.Context, hookCtx *ports.HookContext) error {
	if p == nil || len(p.hooks) == 0 {
		return nil
	}

	for _, h := range p.hooks {
		if err := p.executeSafely(ctx, hookCtx, "Before", func() error {
			return h.Before(ctx, hookCtx)
		}); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteAfter runs all After hooks in FIFO order.
// Returns immediately (short-circuits) if any hook returns an error.
// If a hook panics, the panic is recovered, logged, and converted to an error.
func (p *Pipeline) ExecuteAfter(ctx context.Context, hookCtx *ports.HookContext) error {
	if p == nil || len(p.hooks) == 0 {
		return nil
	}

	for _, h := range p.hooks {
		if err := p.executeSafely(ctx, hookCtx, "After", func() error {
			return h.After(ctx, hookCtx)
		}); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteOnError runs all OnError hooks in FIFO order with the provided error.
// All hooks execute even if some fail; errors from hooks are collected but the
// original error is preserved (hooks cannot mask the original failure).
// If a hook panics, the panic is recovered and logged.
func (p *Pipeline) ExecuteOnError(ctx context.Context, hookCtx *ports.HookContext, err error) error {
	if p == nil || len(p.hooks) == 0 {
		return err
	}

	origErr := err
	for _, h := range p.hooks {
		// OnError hooks all execute even if some fail; we don't short-circuit.
		_ = p.executeSafely(ctx, hookCtx, "OnError", func() error {
			return h.OnError(ctx, hookCtx, origErr)
		})
	}

	// Always return the original error; OnError hooks cannot mask it.
	return origErr
}

// executeSafely wraps hook execution with panic recovery.
// If the hook panics, the panic is recovered, logged, and returned as an error.
// Non-panic errors are returned as-is.
func (p *Pipeline) executeSafely(ctx context.Context, hookCtx *ports.HookContext, phase string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Hook panicked; convert to an error.
			panicMsg := fmt.Sprintf("%v", r)
			stackTrace := string(debug.Stack())

			// Log the panic with full stack trace for debugging.
			// This ensures panics are never silently swallowed.
			if logger, ok := ctx.Value("logger").(ports.Logger); ok {
				logger.Error(
					fmt.Sprintf("hook panic in %s phase for operation %s", phase, hookCtx.Operation),
					fmt.Sprintf("panic value: %s\nstack trace:\n%s", panicMsg, stackTrace),
				)
			}

			// Convert to error return value.
			err = fmt.Errorf("hook panic in %s phase: %v", phase, r)
		}
	}()

	return fn()
}
