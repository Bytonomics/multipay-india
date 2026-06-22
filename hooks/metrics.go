package hooks

import (
	"context"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// MetricsCollector defines an interface for recording hook execution metrics.
type MetricsCollector interface {
	// RecordOperationStarted records the start of a payment operation.
	RecordOperationStarted(ctx context.Context, provider, operation string)

	// RecordOperationSuccess records a successful payment operation.
	// duration is in milliseconds.
	RecordOperationSuccess(ctx context.Context, provider, operation string, duration int64)

	// RecordOperationError records a failed payment operation.
	// duration is in milliseconds.
	RecordOperationError(ctx context.Context, provider, operation string, duration int64, err error)
}

// MetricsHook records metrics for payment operations.
type MetricsHook struct {
	collector MetricsCollector
}

// NewMetricsHook constructs a MetricsHook with the provided MetricsCollector.
func NewMetricsHook(collector MetricsCollector) *MetricsHook {
	return &MetricsHook{
		collector: collector,
	}
}

// Before records the start of the operation.
func (m *MetricsHook) Before(ctx context.Context, hc *ports.HookContext) (context.Context, error) {
	if m.collector != nil {
		m.collector.RecordOperationStarted(ctx, hc.Provider.String(), hc.RequestType)
	}
	return ctx, nil
}

// After records a successful operation.
func (m *MetricsHook) After(ctx context.Context, hc *ports.HookContext) error {
	if m.collector != nil {
		// Note: In a real implementation, duration would be tracked via context values
		// or passed through HookContext. For now, we record with 0 duration.
		m.collector.RecordOperationSuccess(ctx, hc.Provider.String(), hc.RequestType, 0)
	}
	return nil
}

// OnError records a failed operation.
func (m *MetricsHook) OnError(ctx context.Context, hc *ports.HookContext) error {
	if m.collector != nil {
		m.collector.RecordOperationError(ctx, hc.Provider.String(), hc.RequestType, 0, hc.Error)
	}
	return nil
}
