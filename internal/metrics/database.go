// Package metrics - Database metrics are now handled by Bun's OpenTelemetry integration
//
// This file is kept for backward compatibility but the DatabaseMetricsHook
// should no longer be used. Instead, use Bun's bunotel.NewQueryHook() for
// automatic database metrics collection through OpenTelemetry.
//
// The metrics will be automatically exported to Prometheus via the 
// OpenTelemetry Prometheus exporter configured in the main application.

package metrics

import (
	"context"

	"github.com/uptrace/bun"
)

// DatabaseMetricsHook is deprecated - use Bun's OpenTelemetry integration instead
// This is kept for backward compatibility only
type DatabaseMetricsHook struct{}

// NewDatabaseMetricsHook creates a new database metrics hook
// DEPRECATED: Use Bun's bunotel.NewQueryHook() instead
func NewDatabaseMetricsHook() *DatabaseMetricsHook {
	return &DatabaseMetricsHook{}
}

// BeforeQuery is called before query execution
// DEPRECATED: No longer collects metrics - use bunotel instead
func (h *DatabaseMetricsHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	// No-op: metrics now handled by bunotel automatically
	return ctx
}

// AfterQuery is called after query execution
// DEPRECATED: No longer collects metrics - use bunotel instead  
func (h *DatabaseMetricsHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	// No-op: metrics now handled by bunotel automatically
}
