package metrics

import (
	"context"
	"runtime"
	"time"
)

// SystemMetrics provides methods for recording system-level metrics
type SystemMetrics struct{}

// NewSystemMetrics creates a new system metrics instance
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{}
}

// CollectSystemMetrics collects and updates system metrics
func (s *SystemMetrics) CollectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update goroutine count
	GoroutineCount.Set(float64(runtime.NumGoroutine()))

	// Update memory metrics
	MemoryUsage.WithLabelValues("heap_alloc").Set(float64(memStats.HeapAlloc))
	MemoryUsage.WithLabelValues("heap_sys").Set(float64(memStats.HeapSys))
	MemoryUsage.WithLabelValues("heap_idle").Set(float64(memStats.HeapIdle))
	MemoryUsage.WithLabelValues("heap_inuse").Set(float64(memStats.HeapInuse))
	MemoryUsage.WithLabelValues("heap_released").Set(float64(memStats.HeapReleased))
	MemoryUsage.WithLabelValues("stack_inuse").Set(float64(memStats.StackInuse))
	MemoryUsage.WithLabelValues("stack_sys").Set(float64(memStats.StackSys))
	MemoryUsage.WithLabelValues("total_alloc").Set(float64(memStats.TotalAlloc))
	MemoryUsage.WithLabelValues("sys").Set(float64(memStats.Sys))
}

// StartSystemMetricsCollection starts a goroutine that periodically collects system metrics
func (s *SystemMetrics) StartSystemMetricsCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect initial metrics
	s.CollectSystemMetrics()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.CollectSystemMetrics()
		}
	}
}

// RecordGuildCount records the total number of guilds (for compatibility with existing metrics)
func (s *SystemMetrics) RecordGuildCount(shardID string, count float64) {
	TotalGuilds.WithLabelValues(shardID).Set(count)
}
