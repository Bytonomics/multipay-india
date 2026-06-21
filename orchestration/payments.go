package orchestration

import (
	"context"
	"fmt"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/ports"
)

// PaymentService orchestrates payment operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type PaymentService struct {
	resolver  *ports.ProviderRegistry
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
}

// NewPaymentService constructs a PaymentService with required dependencies.
func NewPaymentService(resolver *ports.ProviderRegistry, validator *capabilities.Validator, pipeline *hooks.Pipeline) *PaymentService {
	return &PaymentService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
	}
}

// GetPayment validates input, checks capability, and retrieves a specific payment.
func (s *PaymentService) GetPayment(ctx context.Context, provider domain.Provider, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentFetch); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	adapter, err := s.resolver.Resolve(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetPayment",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	result, err := adapter.GetPayment(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", err)
	}

	return result, nil
}

// ListPayments validates input, checks capability, and retrieves all payments for an order.
func (s *PaymentService) ListPayments(ctx context.Context, provider domain.Provider, req *domain.GetOrderRequest) ([]*domain.Payment, error) {
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
		RequestType: "ListPayments",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	result, err := adapter.ListPayments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		s.pipeline.ExecuteOnError(ctx, hookCtx, err)
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	hookCtx.ResponseData = result
	if err := s.pipeline.ExecuteAfter(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", err)
	}

	return result, nil
}

// CapturePayment validates input, checks capability, and captures an authorized payment.
// This operation is capability-gated and only available on providers that support it (e.g., Razorpay).
func (s *PaymentService) CapturePayment(ctx context.Context, provider domain.Provider, req *domain.CapturePaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentCapture); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CapturePayment",
		RequestData: req,
	}

	if err := s.pipeline.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	// CapturePayment is a method that would be called on the adapter.
	// Since it's not part of the standard PaymentProvider interface,
	// providers that support capture would need to implement a specialized interface.
	// For now, we document this limitation and return an error.
	// Future: Add CapturePayment to PaymentProvider interface or create specialized interfaces per provider.
	hookCtx.Error = domain.ErrUnsupportedCapability
	s.pipeline.ExecuteOnError(ctx, hookCtx, domain.ErrUnsupportedCapability)
	return nil, fmt.Errorf("capture payment is not yet implemented in adapter interface: %w", domain.ErrUnsupportedCapability)
}
