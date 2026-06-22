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

// PaymentService orchestrates payment operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type PaymentService struct {
	resolver  ports.ProviderResolver
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
	clock     ports.Clock
}

// NewPaymentService constructs a PaymentService with required dependencies.
func NewPaymentService(resolver ports.ProviderResolver, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *PaymentService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &PaymentService{
		resolver:  resolver,
		validator: validator,
		pipeline:  pipeline,
		logger:    wrappedLogger,
		clock:     clock,
	}
}

// GetPayment validates input, checks capability, and retrieves a specific payment.
func (s *PaymentService) GetPayment(ctx context.Context, req *domain.GetPaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider := req.Provider

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetPayment",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.GetPayment(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetPayment", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}

// ListPayments validates input, checks capability, and retrieves all payments for an order.
func (s *PaymentService) ListPayments(ctx context.Context, req *domain.ListPaymentsRequest) ([]*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider := req.Provider

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentList)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	adapter, err := s.resolver.Resolve(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve adapter: %w", err)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "ListPayments",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.ListPayments(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for ListPayments", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}

// CapturePayment validates input, checks capability, and captures an authorized payment.
// This operation is capability-gated and only available on providers that support it (e.g., Razorpay).
func (s *PaymentService) CapturePayment(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	provider := req.Provider

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentCapture)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CapturePayment",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, err := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", err)
	}

	// CapturePayment is a method that would be called on the adapter.
	// Since it's not part of the standard PaymentProvider interface,
	// providers that support capture would need to implement a specialized interface.
	// For now, we document this limitation and return an error.
	// Future: Add CapturePayment to PaymentProvider interface or create specialized interfaces per provider.
	hookCtx.Error = domain.ErrUnsupportedCapability
	if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, domain.ErrUnsupportedCapability); hookErr != nil {
		s.logger.Error(ctx, "error in OnError hook for CapturePayment", "error", hookErr.Error())
	}
	return nil, fmt.Errorf("capture payment is not yet implemented in adapter interface: %w", domain.ErrUnsupportedCapability)
}
