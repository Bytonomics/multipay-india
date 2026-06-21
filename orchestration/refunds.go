package orchestration

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// RefundService orchestrates refund operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type RefundService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
}

// NewRefundService constructs a RefundService with required dependencies.
func NewRefundService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline) *RefundService {
	return &RefundService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
	}
}

// CreateRefund validates input, checks capability, and creates a refund for an order.
func (s *RefundService) CreateRefund(ctx context.Context, provider domain.Provider, req *domain.CreateRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapRefundCreate)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreateRefund",
		RequestData: req,
	}

	hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.CreateRefund(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}

// GetRefund validates input, checks capability, and retrieves a specific refund.
func (s *RefundService) GetRefund(ctx context.Context, provider domain.Provider, req *domain.GetRefundRequest) (*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapRefundFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetRefund",
		RequestData: req,
	}

	hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.GetRefund(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}

// ListRefunds validates input, checks capability, and retrieves all refunds for an order.
func (s *RefundService) ListRefunds(ctx context.Context, provider domain.Provider, req *domain.GetOrderRequest) ([]*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapRefundList)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListRefunds",
		RequestData: req,
	}

	hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.ListRefunds(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}
