package metrics

import (
	"context"
	"time"
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

	// Record duration
	SchedulerJobDuration.WithLabelValues(handlerType, status).Observe(duration.Seconds())

	// Record job count
	SchedulerJobTotal.WithLabelValues(handlerType, status).Inc()
}

// RecordActiveJob increments the active job counter for a handler type
func (s *SchedulerMetrics) RecordActiveJob(handlerType string) {
	SchedulerActiveJobs.WithLabelValues(handlerType).Inc()
}

// RecordJobComplete decrements the active job counter for a handler type
func (s *SchedulerMetrics) RecordJobComplete(handlerType string) {
	SchedulerActiveJobs.WithLabelValues(handlerType).Dec()
}

// UpdateQueueDepth updates the scheduler queue depth metric
func (s *SchedulerMetrics) UpdateQueueDepth(depth float64) {
	SchedulerQueueDepth.Set(depth)
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
