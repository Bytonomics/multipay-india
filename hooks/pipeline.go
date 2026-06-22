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
	hooks  []ports.Hook
	logger ports.Logger
}

// NewPipeline constructs a Pipeline with an ordered slice of hooks.
// Hooks will execute in the order provided.
func NewPipeline(logger ports.Logger, hooks ...ports.Hook) *Pipeline {
	return &Pipeline{
		hooks:  hooks,
		logger: logger,
	}
}

// ExecuteBefore runs all Before hooks in FIFO order.
// Returns the modified context and immediately (short-circuits) if any hook returns an error.
// If a hook panics, the panic is recovered, logged, and converted to an error.
func (p *Pipeline) ExecuteBefore(ctx context.Context, hookCtx *ports.HookContext) (context.Context, error) {
	if p == nil || len(p.hooks) == 0 {
		return ctx, nil
	}
	return p.executeBeforeHooks(ctx, hookCtx)
}

// executeBeforeHooks is a helper that threads context through hooks without named returns.
func (p *Pipeline) executeBeforeHooks(ctx context.Context, hookCtx *ports.HookContext) (context.Context, error) {
	newCtx := ctx
	for _, h := range p.hooks {
		var err error
		newCtx, err = p.beforeWithPanicRecovery(ctx, hookCtx, h)
		if err != nil {
			return nil, err
		}
	}
	return newCtx, nil
}

// ExecuteAfter runs all After hooks in LIFO order.
// Returns immediately (short-circuits) if any hook returns an error.
// If a hook panics, the panic is recovered, logged, and converted to an error.
func (p *Pipeline) ExecuteAfter(ctx context.Context, hookCtx *ports.HookContext) error {
	if p == nil || len(p.hooks) == 0 {
		return nil
	}

	for i := len(p.hooks) - 1; i >= 0; i-- {
		h := p.hooks[i]
		if err := p.afterWithPanicRecovery(ctx, hookCtx, h); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteOnError runs all OnError hooks in LIFO order with the provided error.
// Returns an error if any hook fails; errors from hooks are logged but not propagated beyond return.
// If a hook panics, the panic is recovered and logged.
// OnError hooks cannot mask or modify the original error; they run for side effects only.
func (p *Pipeline) ExecuteOnError(ctx context.Context, hookCtx *ports.HookContext, err error) error {
	if p == nil || len(p.hooks) == 0 {
		return nil
	}

	// Set the error in hookCtx so hooks have access to it.
	hookCtx.Error = err

	for i := len(p.hooks) - 1; i >= 0; i-- {
		h := p.hooks[i]
		// OnError hooks all execute even if some fail; we don't short-circuit.
		hookErr := p.onErrorWithPanicRecovery(ctx, hookCtx, h)
		if hookErr != nil {
			// Log hook errors for debugging, but don't propagate them.
			p.logger.Error(ctx,
				"error in OnError hook for operation "+hookCtx.RequestType,
				"error", hookErr.Error(),
			)
		}
	}
	return nil
}

// beforeWithPanicRecovery executes a Before hook with panic recovery.
func (p *Pipeline) beforeWithPanicRecovery(ctx context.Context, hookCtx *ports.HookContext, h ports.Hook) (context.Context, error) {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			p.logger.Error(ctx,
				"hook panic in Before phase for operation "+hookCtx.RequestType,
				"panic", fmt.Sprintf("%v", r), "stacktrace", stackTrace,
			)
		}
	}()

	resCtx, err := h.Before(ctx, hookCtx)
	if err != nil {
		return nil, fmt.Errorf("before hook: %w", err)
	}
	return resCtx, nil
}

// afterWithPanicRecovery executes an After hook with panic recovery.
func (p *Pipeline) afterWithPanicRecovery(ctx context.Context, hookCtx *ports.HookContext, h ports.Hook) error {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			p.logger.Error(ctx,
				"hook panic in After phase for operation "+hookCtx.RequestType,
				"panic", fmt.Sprintf("%v", r), "stacktrace", stackTrace,
			)
		}
	}()

	err := h.After(ctx, hookCtx)
	if err != nil {
		return fmt.Errorf("after hook: %w", err)
	}
	return nil
}

// onErrorWithPanicRecovery executes an OnError hook with panic recovery.
func (p *Pipeline) onErrorWithPanicRecovery(ctx context.Context, hookCtx *ports.HookContext, h ports.Hook) error {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			p.logger.Error(ctx,
				"hook panic in OnError phase for operation "+hookCtx.RequestType,
				"panic", fmt.Sprintf("%v", r), "stacktrace", stackTrace,
			)
		}
	}()

	err := h.OnError(ctx, hookCtx)
	if err != nil {
		return fmt.Errorf("on-error hook: %w", err)
	}
	return nil
}
