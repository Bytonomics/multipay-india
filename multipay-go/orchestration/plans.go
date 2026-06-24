package orchestration

import (
	"context"
	"fmt"

	"github.com/SmrutAI/pedantigo"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/hooks"
	"github.com/Bytonomics/multipay-india/multipay-go/logging"
	"github.com/Bytonomics/multipay-india/multipay-go/ports"
)

var (
	createPlanValidator = pedantigo.New[domain.CreatePlanRequest]()
	getPlanValidator    = pedantigo.New[domain.GetPlanRequest]()
)

// PlanService orchestrates plan operations across multiple payment providers.
// Plans are FIRST-CLASS — both Cashfree and Razorpay support CreatePlan and GetPlan.
// No capability validation is required (unlike capability-gated services).
type PlanService struct {
	adapter  ports.ProviderAdapter
	provider domain.Provider
	pipeline *hooks.Pipeline
	logger   ports.Logger
	clock    ports.Clock
}

// NewPlanService constructs a PlanService with required dependencies.
// Logger is mandatory and will panic if nil.
func NewPlanService(provider domain.Provider, adapter ports.ProviderAdapter, pipeline *hooks.Pipeline, logger ports.Logger, clock ports.Clock) *PlanService {
	if logger == nil {
		panic("logger is required (cannot be nil)")
	}
	wrappedLogger := logging.NewCallerLogger(logger, 2)

	return &PlanService{
		adapter:  adapter,
		provider: provider,
		pipeline: pipeline,
		logger:   wrappedLogger,
		clock:    clock,
	}
}

// CreatePlan validates input, executes hooks, and creates a plan.
// This is a first-class operation — no capability gate.
func (s *PlanService) CreatePlan(ctx context.Context, req *domain.CreatePlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := createPlanValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "CreatePlan",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.CreatePlan(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for CreatePlan", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("create plan failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}

// GetPlan validates input, executes hooks, and retrieves a plan.
// This is a first-class operation — no capability gate.
func (s *PlanService) GetPlan(ctx context.Context, req *domain.GetPlanRequest) (*domain.Plan, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil: %w", domain.ErrInvalidRequest)
	}

	if err := getPlanValidator.Validate(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	provider := s.provider
	adapter := s.adapter

	hookCtx := &ports.HookContext{
		Provider:    provider,
		RequestType: "GetPlan",
		RequestData: req,
		StartTime:   s.clock.Now(),
	}

	ctx, hookErr := s.pipeline.ExecuteBefore(ctx, hookCtx)
	if hookErr != nil {
		return nil, fmt.Errorf("before hook failed: %w", hookErr)
	}

	result, err := adapter.GetPlan(ctx, req)
	if err != nil {
		hookCtx.Error = err
		if hookErr := s.pipeline.ExecuteOnError(ctx, hookCtx, err); hookErr != nil {
			s.logger.Error(ctx, "error in OnError hook for GetPlan", "error", hookErr.Error())
		}
		return nil, fmt.Errorf("get plan failed: %w", err)
	}

	hookCtx.ResponseData = result
	afterErr := s.pipeline.ExecuteAfter(ctx, hookCtx)
	if afterErr != nil {
		return nil, fmt.Errorf("after hook failed: %w", afterErr)
	}

	return result, nil
}
