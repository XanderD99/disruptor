package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// SystemMetrics provides methods for recording system-level metrics
type SystemMetrics struct{}

// NewSystemMetrics creates a new system metrics instance
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{}
}

// CollectSystemMetrics is no longer needed as we use observable gauges
// The system metrics are automatically collected via OpenTelemetry callbacks
func (s *SystemMetrics) CollectSystemMetrics() {
	// This method is kept for backward compatibility but does nothing
	// System metrics are now collected automatically via observable gauges
}

// StartSystemMetricsCollection is no longer needed with observable gauges
// but kept for backward compatibility
func (s *SystemMetrics) StartSystemMetricsCollection(ctx context.Context, interval time.Duration) {
	// Observable gauges in OpenTelemetry automatically collect metrics
	// when the metrics are scraped, so no active collection is needed.
	// This method will just wait for context cancellation to maintain
	// backward compatibility with existing code that expects this method.
	<-ctx.Done()
}

// RecordGuildCount records the total number of guilds (for compatibility with existing metrics)
func (s *SystemMetrics) RecordGuildCount(shardID string, count int64) {
	ctx := context.Background()
	TotalGuilds.Add(ctx, count, metric.WithAttributes(
		attribute.String("shard", shardID),
	))
}
