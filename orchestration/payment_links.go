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

// PaymentLinkService orchestrates payment link operations with validation, capability checking, and hooks.
type PaymentLinkService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
}

// NewPaymentLinkService constructs a PaymentLinkService with dependency injection.
func NewPaymentLinkService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline) *PaymentLinkService {
	return &PaymentLinkService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
	}
}

// CreatePaymentLink creates a new shareable payment link with validation and hook execution.
func (s *PaymentLinkService) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	if err := s.validator.RequireCapability(ctx, provider, domain.CapPaymentLinkCreate); err != nil {
		return nil, err
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreatePaymentLink",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("before hook failed: %w", err)
	}

	result, err := adapter.CreatePaymentLink(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("create payment link failed: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("after hook failed: %w", err)
	}

	return result, nil
}

// GetPaymentLink retrieves an existing payment link with validation and hook execution.
func (s *PaymentLinkService) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	if err := s.validator.RequireCapability(ctx, provider, domain.CapPaymentLinkFetch); err != nil {
		return nil, err
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetPaymentLink",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("before hook failed: %w", err)
	}

	result, err := adapter.GetPaymentLink(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("get payment link failed: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("after hook failed: %w", err)
	}

	return result, nil
}

// CancelPaymentLink cancels an existing payment link with validation and hook execution.
func (s *PaymentLinkService) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLinkResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider, ok := ctx.Value("provider").(domain.Provider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}

	if err := s.validator.RequireCapability(ctx, provider, domain.CapPaymentLinkCancel); err != nil {
		return nil, err
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("provider resolution failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CancelPaymentLink",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("before hook failed: %w", err)
	}

	result, err := adapter.CancelPaymentLink(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("cancel payment link failed: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("after hook failed: %w", err)
	}

	return result, nil
}
