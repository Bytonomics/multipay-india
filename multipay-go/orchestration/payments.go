package orchestration

import (
	"context"
	"fmt"

	"github.com/SmrutAI/pedantigo"

	"github.com/Bytonomics/multipay-india/multipay-go/capabilities"
	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/logging"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

var (
	getPaymentValidator     = pedantigo.New[domain.GetPaymentRequest]()
	listPaymentsValidator   = pedantigo.New[domain.ListPaymentsRequest]()
	capturePaymentValidator = pedantigo.New[domain.CapturePaymentRequest]()
)

// PaymentService orchestrates payment operations across multiple payment providers.
// It handles validation, capability checking, and hook execution.
type PaymentService struct {
	adapter   ports.ProviderAdapter
	provider  domain.Provider
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
	clock     ports.Clock
}

// NewPaymentService constructs a PaymentService with required dependencies.
func NewPaymentService(provider domain.Provider, adapter ports.ProviderAdapter, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *PaymentService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &PaymentService{
		adapter:   adapter,
		provider:  provider,
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

	if err := getPaymentValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
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

	if err := listPaymentsValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentList)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
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
func (s *PaymentService) CapturePayment(ctx context.Context, req *domain.CapturePaymentRequest) (*domain.Payment, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := capturePaymentValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

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

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("failed to execute before hooks: %w", hookErr)
	}

	result, err := adapter.CapturePayment(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if onErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); onErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CapturePayment", "error", onErr.Error())
		}
		return nil, fmt.Errorf("failed to capture payment: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("failed to execute after hooks: %w", afterErr)
	}

	return result, nil
}
