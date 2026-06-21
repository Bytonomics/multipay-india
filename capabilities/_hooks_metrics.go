// This file is temporarily placed in capabilities/ but should be moved to hooks/metrics.go
package hooks

import (
	"context"
	"fmt"
	"time"

	"github.com/Bytonomics/multipay-adapter/ports"
)

// Metric name constants.
const (
	MetricOperationCount    = "multipay.operation.count"
	MetricOperationDuration = "multipay.operation.duration_ms"
	MetricOperationErrors   = "multipay.operation.errors"
)

// MetricsCollector defines the interface for collecting operation metrics.
// Implementations can record counters, histograms, and error counts.
type MetricsCollector interface {
	// RecordCounter increments a counter metric by the specified amount.
	RecordCounter(ctx context.Context, metricName string, value int64, tags map[string]string)

	// RecordHistogram records a histogram value for a metric.
	RecordHistogram(ctx context.Context, metricName string, value float64, tags map[string]string)

	// RecordError increments an error counter for a metric.
	RecordError(ctx context.Context, metricName string, tags map[string]string)
}

// MetricsHook collects operation metrics (counters, durations, error counts).
// It implements the ports.Hook interface.
type MetricsHook struct {
	collector  MetricsCollector
	startTimes map[string]time.Time // Keyed by operation context
}

// NewMetricsHook constructs a MetricsHook with a metrics collector.
// The collector must not be nil.
func NewMetricsHook(collector MetricsCollector) *MetricsHook {
	return &MetricsHook{
		collector:  collector,
		startTimes: make(map[string]time.Time),
	}
}

// Before records the operation start time and increments the operation counter.
func (h *MetricsHook) Before(ctx context.Context, hookCtx *ports.HookContext) error {
	if h.collector == nil {
		return nil
	}

	// Record operation start counter.
	tags := h.tagsForHook(hookCtx)
	h.collector.RecordCounter(ctx, MetricOperationCount, 1, tags)

	// Track start time for duration calculation in After/OnError.
	contextKey := h.contextKey(hookCtx)
	h.startTimes[contextKey] = time.Now()

	return nil
}

// After records the operation duration histogram.
func (h *MetricsHook) After(ctx context.Context, hookCtx *ports.HookContext) error {
	if h.collector == nil {
		return nil
	}

	// Calculate and record duration.
	contextKey := h.contextKey(hookCtx)
	if startTime, ok := h.startTimes[contextKey]; ok {
		duration := time.Since(startTime)
		durationMs := float64(duration.Milliseconds())

		tags := h.tagsForHook(hookCtx)
		h.collector.RecordHistogram(ctx, MetricOperationDuration, durationMs, tags)

		// Clean up start time tracking.
		delete(h.startTimes, contextKey)
	}

	return nil
}

// OnError records an error counter and the operation duration.
func (h *MetricsHook) OnError(ctx context.Context, hookCtx *ports.HookContext, err error) error {
	if h.collector == nil {
		return nil
	}

	tags := h.tagsForHook(hookCtx)

	// Record error counter.
	h.collector.RecordError(ctx, MetricOperationErrors, tags)

	// Record duration even on error.
	contextKey := h.contextKey(hookCtx)
	if startTime, ok := h.startTimes[contextKey]; ok {
		duration := time.Since(startTime)
		durationMs := float64(duration.Milliseconds())
		h.collector.RecordHistogram(ctx, MetricOperationDuration, durationMs, tags)

		// Clean up start time tracking.
		delete(h.startTimes, contextKey)
	}

	return nil
}

// tagsForHook returns a tag map for the operation and provider.
func (h *MetricsHook) tagsForHook(hookCtx *ports.HookContext) map[string]string {
	return map[string]string{
		"operation": hookCtx.Operation,
		"provider":  string(hookCtx.Provider),
	}
}

// contextKey returns a unique key for the hook context (for tracking start times).
func (h *MetricsHook) contextKey(hookCtx *ports.HookContext) string {
	return fmt.Sprintf("%s:%s", hookCtx.Operation, hookCtx.Provider)
}

// Compile-time check: MetricsHook implements ports.Hook.
var _ ports.Hook = (*MetricsHook)(nil)
