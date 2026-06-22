package orchestration

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/logging"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// RefundService orchestrates refund operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type RefundService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
}

// NewRefundService constructs a RefundService with required dependencies.
func NewRefundService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger) *RefundService {
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &RefundService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
		logger:    wrappedLogger,
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

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreateRefund",
		RequestData: req,
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.CreateRefund(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CreateRefund", "error", hookErr.Error())
		}
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

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetRefund",
		RequestData: req,
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.GetRefund(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetRefund", "error", hookErr.Error())
		}
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
func (s *RefundService) ListRefunds(ctx context.Context, provider domain.Provider, req *domain.ListRefundsRequest) ([]*domain.Refund, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapRefundList)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListRefunds",
		RequestData: req,
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.ListRefunds(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ListRefunds", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}
