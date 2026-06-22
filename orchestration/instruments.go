package orchestration

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/logging"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// InstrumentService orchestrates instrument operations with validation, capability checking, and hooks.
type InstrumentService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
}

// NewInstrumentService constructs an InstrumentService with dependency injection.
func NewInstrumentService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger) *InstrumentService {
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &InstrumentService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
		logger:    wrappedLogger,
	}
}

// GetInstrument retrieves a specific payment instrument with validation and hook execution.
func (s *InstrumentService) GetInstrument(ctx context.Context, req *domain.GetInstrumentRequest) (*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapInstrumentFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetInstrument",
		RequestData: req,
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.GetInstrument(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetInstrument", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("get instrument failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// ListInstruments retrieves all instruments for a customer with validation and hook execution.
func (s *InstrumentService) ListInstruments(ctx context.Context, req *domain.ListInstrumentsRequest) ([]*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapInstrumentList)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListInstruments",
		RequestData: req,
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.ListInstruments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ListInstruments", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("list instruments failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// DeleteInstrument removes a payment instrument with validation and hook execution.
func (s *InstrumentService) DeleteInstrument(ctx context.Context, req *domain.DeleteInstrumentRequest) (*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapInstrumentDelete)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "DeleteInstrument",
		RequestData: req,
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, deleteErr := adapter.DeleteInstrument(ctx, req)
	if deleteErr != nil {
		hookCtx.Error = deleteErr
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, deleteErr); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for DeleteInstrument", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("delete instrument failed: %w", deleteErr)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}
