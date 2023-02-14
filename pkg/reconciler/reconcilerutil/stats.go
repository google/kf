package reconcilerutil

import (
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/controller"
)

var (
	reportFrequency = 30 * time.Second
)

// StructuredStatsReporter logs events
type StructuredStatsReporter struct {
	// Logger to write structured output to.
	// Must be set before Report functions are called.
	Logger *zap.SugaredLogger

	// Clock is an optional clock.
	Clock func() time.Time

	// Reporting data.
	lastReportNanos atomic.Int64

	// Queue monitoring data.
	currentQueueDepth atomic.Int64

	// Reconcliation monitoring data.
	successfulReconciliations atomic.Uint64
	failedReconciliations     atomic.Uint64
	unknownReconciliations    atomic.Uint64
	reconciliationTimeMS      atomic.Int64
}

var _ controller.StatsReporter = (*StructuredStatsReporter)(nil)

// now uses the system clock or the overridden clock
func (s *StructuredStatsReporter) now() time.Time {
	if s.Clock == nil {
		return time.Now()
	}

	return s.Clock()
}

// ReportQueueDepth reports the queue depth metric
func (s *StructuredStatsReporter) ReportQueueDepth(v int64) error {
	s.currentQueueDepth.Store(v)

	s.MaybeReport()

	return nil
}

// ReportReconcile reports the count and latency metrics for a reconcile operation
func (s *StructuredStatsReporter) ReportReconcile(duration time.Duration, success string, key types.NamespacedName) error {
	s.reconciliationTimeMS.Add(duration.Milliseconds())

	switch success {
	case "true":
		s.successfulReconciliations.Add(1)
	case "false":
		s.failedReconciliations.Add(1)
	default:
		s.unknownReconciliations.Add(1)
	}

	s.MaybeReport()

	return nil
}

// MaybeReport writes a report to the logger if it hasn't been done recently.
// Returns true if the report was triggered and the internal state was updated.
func (s *StructuredStatsReporter) MaybeReport() bool {
	// The StatsReporter interface doesn't provide any guarantees about how
	// ReportXXX functions will be called, so it's safest to assume multiple
	// threads might be calling simultaneously. Atomics are safer because
	// they won't deadlock even if they have a small window where race conditions
	// can happen.

	// Rate limit the reporting.
	currentTime := s.now()
	lastReportNanos := s.lastReportNanos.Load()
	currentNanos := currentTime.UnixNano()
	reportWindow := time.Duration(currentNanos - lastReportNanos)

	if reportWindow < reportFrequency {
		// Last report was too recent.
		return false
	}

	if !s.lastReportNanos.CompareAndSwap(lastReportNanos, currentNanos) {
		// A different thread is reporting.
		return false
	}

	currentQueueDepth := s.currentQueueDepth.Load()
	successfulReconciliations := s.successfulReconciliations.Swap(0)
	failedReconciliations := s.failedReconciliations.Swap(0)
	unknownReconciliations := s.unknownReconciliations.Swap(0)
	timeSpentReconcilingMS := s.reconciliationTimeMS.Swap(0)

	totalReconciliations := successfulReconciliations + failedReconciliations + unknownReconciliations

	var averageMSPerReconciliation float64
	var reconciliationsPerSecond float64
	if totalReconciliations > 0 {
		averageMSPerReconciliation = float64(timeSpentReconcilingMS) / float64(totalReconciliations)
		reconciliationsPerSecond = float64(totalReconciliations) / reportWindow.Seconds()
	}

	if logger := s.Logger; logger != nil {
		logger.With(
			"currentQueueDepth", currentQueueDepth,
			"averageMSPerReconciliation", averageMSPerReconciliation,
			"reconciliationsPerSecond", reconciliationsPerSecond,
			"statuses", map[string]uint64{
				"success": successfulReconciliations,
				"failure": failedReconciliations,
				"unknown": unknownReconciliations,
			},
		).Info("Controller Statistics")
	}

	return true
}
