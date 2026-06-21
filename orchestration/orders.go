package orchestration

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// OrderService orchestrates order operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type OrderService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
}

// NewOrderService constructs an OrderService with required dependencies.
func NewOrderService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline) *OrderService {
	return &OrderService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
	}
}

// CreateOrder validates input, checks capability, and creates an order on the payment provider.
func (s *OrderService) CreateOrder(ctx context.Context, provider domain.Provider, req *domain.CreateOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapOrderCreate); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreateOrder",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	result, err := adapter.CreateOrder(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", err)
	}

	return result, nil
}

// GetOrder validates input, checks capability, and retrieves an order from the payment provider.
func (s *OrderService) GetOrder(ctx context.Context, provider domain.Provider, req *domain.GetOrderRequest) (*domain.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapOrderFetch); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetOrder",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	result, err := adapter.GetOrder(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", err)
	}

	return result, nil
}

// ListOrderPayments validates input, checks capability, and retrieves all payments for an order.
func (s *OrderService) ListOrderPayments(ctx context.Context, provider domain.Provider, req *domain.GetOrderRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentList); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListOrderPayments",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	result, err := adapter.ListOrderPayments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to list order payments: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", err)
	}

	return result, nil
}
