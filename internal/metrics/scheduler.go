package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// SchedulerMetrics provides methods for recording scheduler-related metrics
type SchedulerMetrics struct{}

// NewSchedulerMetrics creates a new scheduler metrics instance
func NewSchedulerMetrics() *SchedulerMetrics {
	return &SchedulerMetrics{}
}

// RecordJobExecution records metrics for a scheduler job execution
func (s *SchedulerMetrics) RecordJobExecution(handlerType string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	ctx := context.Background()

	// Record duration
	SchedulerJobDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("handler_type", handlerType),
		attribute.String("status", status),
	))

	// Record job count
	SchedulerJobTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("handler_type", handlerType),
		attribute.String("status", status),
	))
}

// RecordActiveJob increments the active job counter for a handler type
func (s *SchedulerMetrics) RecordActiveJob(handlerType string) {
	ctx := context.Background()
	SchedulerActiveJobs.Add(ctx, 1, metric.WithAttributes(
		attribute.String("handler_type", handlerType),
	))
}

// RecordJobComplete decrements the active job counter for a handler type
func (s *SchedulerMetrics) RecordJobComplete(handlerType string) {
	ctx := context.Background()
	SchedulerActiveJobs.Add(ctx, -1, metric.WithAttributes(
		attribute.String("handler_type", handlerType),
	))
}

// UpdateQueueDepth updates the scheduler queue depth metric
func (s *SchedulerMetrics) UpdateQueueDepth(depth int64) {
	ctx := context.Background()
	SchedulerQueueDepth.Add(ctx, depth)
}

// JobExecutionTimer provides a timer for measuring job execution duration
type JobExecutionTimer struct {
	metrics     *SchedulerMetrics
	handlerType string
	startTime   time.Time
}

// NewJobExecutionTimer creates a new timer for job execution
func (s *SchedulerMetrics) NewJobExecutionTimer(handlerType string) *JobExecutionTimer {
	s.RecordActiveJob(handlerType)
	return &JobExecutionTimer{
		metrics:     s,
		handlerType: handlerType,
		startTime:   time.Now(),
	}
}

// Finish completes the timer and records the execution metrics
func (t *JobExecutionTimer) Finish(err error) {
	duration := time.Since(t.startTime)
	t.metrics.RecordJobExecution(t.handlerType, duration, err)
	t.metrics.RecordJobComplete(t.handlerType)
}

// WithJobMetrics is a helper function that wraps a scheduler handler function
// with automatic metrics collection
func WithJobMetrics(handlerType string, handler func(ctx context.Context) error) func(ctx context.Context) error {
	metrics := NewSchedulerMetrics()

	return func(ctx context.Context) error {
		timer := metrics.NewJobExecutionTimer(handlerType)
		err := handler(ctx)
		timer.Finish(err)
		return err
	}
}
