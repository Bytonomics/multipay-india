package orchestration

import (
	"context"
	"fmt"

	"github.com/SmrutAI/pedantigo"

	"github.com/Bytonomics/multipay-adapter/capabilities"
	"github.com/Bytonomics/multipay-adapter/domain"
	"github.com/Bytonomics/multipay-adapter/hooks"
	"github.com/Bytonomics/multipay-adapter/logging"
	"github.com/Bytonomics/multipay-adapter/ports"
)

var (
	createPaymentLinkValidator = pedantigo.New[domain.CreatePaymentLinkRequest]()
	getPaymentLinkValidator    = pedantigo.New[domain.GetPaymentLinkRequest]()
	cancelPaymentLinkValidator = pedantigo.New[domain.CancelPaymentLinkRequest]()
)

// PaymentLinkService orchestrates payment link operations with validation, capability checking, and hooks.
type PaymentLinkService struct {
	adapter   ports.ProviderAdapter
	provider  domain.Provider
	validator *capabilities.Validator
	pipeline  *hooks.Pipeline
	logger    ports.Logger
	clock     ports.Clock
}

// NewPaymentLinkService constructs a PaymentLinkService with dependency injection.
func NewPaymentLinkService(provider domain.Provider, adapter ports.ProviderAdapter, validator *capabilities.Validator, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *PaymentLinkService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &PaymentLinkService{
		adapter:   adapter,
		provider:  provider,
		validator: validator,
		pipeline:  pipeline,
		logger:    wrappedLogger,
		clock:     clock,
	}
}

// CreatePaymentLink creates a new shareable payment link with validation and hook execution.
func (s *PaymentLinkService) CreatePaymentLink(ctx context.Context, req *domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := createPaymentLinkValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentLinkCreate)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreatePaymentLink",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.CreatePaymentLink(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CreatePaymentLink", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("create payment link failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// GetPaymentLink retrieves an existing payment link with validation and hook execution.
func (s *PaymentLinkService) GetPaymentLink(ctx context.Context, req *domain.GetPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := getPaymentLinkValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentLinkFetch)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetPaymentLink",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.GetPaymentLink(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetPaymentLink", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("get payment link failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// CancelPaymentLink cancels an existing payment link with validation and hook execution.
func (s *PaymentLinkService) CancelPaymentLink(ctx context.Context, req *domain.CancelPaymentLinkRequest) (*domain.PaymentLink, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := cancelPaymentLinkValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	capErr := s.validator.RequireCapability(ctx, provider, capabilities.CapPaymentLinkCancel)
	if capErr != nil {
		return nil, fmt.Errorf("capability check failed: %w", capErr)
	}

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CancelPaymentLink",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.CancelPaymentLink(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CancelPaymentLink", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("cancel payment link failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}
