package metrics

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// DatabaseMetricsHook implements bun.QueryHook to collect database metrics
type DatabaseMetricsHook struct{}

// NewDatabaseMetricsHook creates a new database metrics hook
func NewDatabaseMetricsHook() *DatabaseMetricsHook {
	return &DatabaseMetricsHook{}
}

// BeforeQuery is called before query execution
func (h *DatabaseMetricsHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	// Store start time in context for duration calculation
	return context.WithValue(ctx, "metrics_start_time", time.Now())
}

// AfterQuery is called after query execution and collects metrics
func (h *DatabaseMetricsHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	// Calculate duration
	startTime, ok := ctx.Value("metrics_start_time").(time.Time)
	if !ok {
		startTime = event.StartTime
	}
	duration := time.Since(startTime).Seconds()

	// Extract operation and table information
	operation := h.getOperation(event)
	table := h.getTable(event)

	// Record duration histogram
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)

	// Record query count
	status := "success"
	if event.Err != nil {
		status = "error"

		// Record error metrics
		errorType := h.getErrorType(event.Err)
		DatabaseErrors.WithLabelValues(operation, table, errorType).Inc()
	}

	DatabaseQueryTotal.WithLabelValues(operation, table, status).Inc()
}

// getOperation extracts the operation type from the query event
func (h *DatabaseMetricsHook) getOperation(event *bun.QueryEvent) string {
	switch event.QueryAppender.(type) {
	case *bun.SelectQuery:
		return "select"
	case *bun.InsertQuery:
		return "insert"
	case *bun.UpdateQuery:
		return "update"
	case *bun.DeleteQuery:
		return "delete"
	case *bun.CreateTableQuery:
		return "create_table"
	case *bun.DropTableQuery:
		return "drop_table"
	default:
		// Fall back to parsing the query string
		return h.parseOperationFromQuery(string(event.Query))
	}
}

// parseOperationFromQuery extracts operation from raw query string
func (h *DatabaseMetricsHook) parseOperationFromQuery(query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return "unknown"
	}

	// Find the first word which should be the operation
	parts := strings.Fields(query)
	if len(parts) == 0 {
		return "unknown"
	}

	operation := parts[0]
	if len(operation) > 16 {
		operation = operation[:16]
	}

	return operation
}

// getTable attempts to extract table name from the query event
func (h *DatabaseMetricsHook) getTable(event *bun.QueryEvent) string {
	// Try to get table from query appender
	switch q := event.QueryAppender.(type) {
	case *bun.SelectQuery:
		if q.GetTableName() != "" {
			return q.GetTableName()
		}
	case *bun.InsertQuery:
		if q.GetTableName() != "" {
			return q.GetTableName()
		}
	case *bun.UpdateQuery:
		if q.GetTableName() != "" {
			return q.GetTableName()
		}
	case *bun.DeleteQuery:
		if q.GetTableName() != "" {
			return q.GetTableName()
		}
	case *bun.CreateTableQuery:
		if q.GetTableName() != "" {
			return q.GetTableName()
		}
	case *bun.DropTableQuery:
		if q.GetTableName() != "" {
			return q.GetTableName()
		}
	}

	// Fall back to parsing from query string
	return h.parseTableFromQuery(string(event.Query))
}

// parseTableFromQuery attempts to extract table name from raw SQL
func (h *DatabaseMetricsHook) parseTableFromQuery(query string) string {
	query = strings.TrimSpace(strings.ToLower(query))

	// Simple heuristics for common patterns
	if strings.Contains(query, "guilds") {
		return "guilds"
	}

	// Could add more sophisticated parsing here if needed
	return "unknown"
}

// getErrorType categorizes database errors
func (h *DatabaseMetricsHook) getErrorType(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	// Categorize common error types
	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "syntax"):
		return "syntax"
	case strings.Contains(errStr, "constraint"):
		return "constraint"
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no rows"):
		return "not_found"
	case strings.Contains(errStr, "deadlock"):
		return "deadlock"
	default:
		return "other"
	}
}
