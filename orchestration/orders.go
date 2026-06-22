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

// OrderService orchestrates order operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type OrderService struct {
	adapter   ports.ProviderAdapter
	provider  domain.Provider
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
	clock     ports.Clock
}

// NewOrderService constructs an OrderService with required dependencies.
func NewOrderService(provider domain.Provider, adapter ports.ProviderAdapter, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *OrderService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &OrderService{
		adapter:   adapter,
		provider:  provider,
		validator: validator,
		pipeline:  pipeline,
		logger:    wrappedLogger,
		clock:     clock,
	}
}

// CreateOrder validates input, checks capability, and creates an order on the payment provider.
func (s *OrderService) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider := s.provider
	adapter := s.adapter

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapOrderCreate); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreateOrder",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.CreateOrder(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CreateOrder", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}

// GetOrder validates input, checks capability, and retrieves an order from the payment provider.
func (s *OrderService) GetOrder(ctx context.Context, req *domain.GetOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapOrderFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetOrder",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.GetOrder(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetOrder", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}

// ListOrderPayments validates input, checks capability, and retrieves all payments for an order.
func (s *OrderService) ListOrderPayments(ctx context.Context, req *domain.ListOrderPaymentsRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapOrderListPayments)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListOrderPayments",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.ListOrderPayments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ListOrderPayments", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to list order payments: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}
