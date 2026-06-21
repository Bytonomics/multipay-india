package orchestration

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// InstrumentService orchestrates instrument operations with validation, capability checking, and hooks.
type InstrumentService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
}

// NewInstrumentService constructs an InstrumentService with dependency injection.
func NewInstrumentService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline) *InstrumentService {
	return &InstrumentService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
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

	capErr := s.validator.RequireCapability(ctx, provider, domain.CapInstrumentFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetInstrument",
		RequestData: req,
	}

	hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.GetInstrument(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
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
func (s *InstrumentService) ListInstruments(ctx context.Context, req *domain.GetInstrumentRequest) ([]*domain.Instrument, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	capErr := s.validator.RequireCapability(ctx, provider, domain.CapInstrumentList)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListInstruments",
		RequestData: req,
	}

	hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.ListInstruments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
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
func (s *InstrumentService) DeleteInstrument(ctx context.Context, req *domain.GetInstrumentRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return errors.New("provider not found in context")
	}

	capErr := s.validator.RequireCapability(ctx, provider, domain.CapInstrumentDelete)
	if capErr != nil {
		return fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "DeleteInstrument",
		RequestData: req,
	}

	hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return fmt.Errorf("before hook failed: %w", hookErr)
	}

	deleteErr := adapter.DeleteInstrument(ctx, req)
	if deleteErr != nil {
		hookCtx.Error = deleteErr
		s.pipeline.ExecuteOnError(ctx, hookCtx, deleteErr)
		return fmt.Errorf("delete instrument failed: %w", deleteErr)
	}

	hookCtx.ResponseData = map[string]interface{}{"deleted": true}
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return fmt.Errorf("after hook failed: %w", afterErr)
	}

	return nil
}
