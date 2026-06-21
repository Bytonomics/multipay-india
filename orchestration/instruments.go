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

	if err := s.validator.RequireCapability(ctx, provider, domain.CapInstrumentFetch); err != nil {
		return nil, err
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

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("before hook failed: %w", err)
	}

	result, err := adapter.GetInstrument(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("get instrument failed: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("after hook failed: %w", err)
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

	if err := s.validator.RequireCapability(ctx, provider, domain.CapInstrumentList); err != nil {
		return nil, err
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

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("before hook failed: %w", err)
	}

	result, err := adapter.ListInstruments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("list instruments failed: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("after hook failed: %w", err)
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

	if err := s.validator.RequireCapability(ctx, provider, domain.CapInstrumentDelete); err != nil {
		return err
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

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return fmt.Errorf("before hook failed: %w", err)
	}

	err = adapter.DeleteInstrument(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return fmt.Errorf("delete instrument failed: %w", err)
	}

	hookCtx.ResponseData = map[string]interface{}{"deleted": true}
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return fmt.Errorf("after hook failed: %w", err)
	}

	return nil
}
