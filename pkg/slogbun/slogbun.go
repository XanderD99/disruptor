package slogbun

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/uptrace/bun"
)

type Option func(hook *QueryHook)

// WithEnabled enables/disables this hook
func WithEnabled(on bool) Option {
	return func(h *QueryHook) {
		h.enabled = on
	}
}

// WithVerbose configures the hook to log all queries
// (by default, only failed queries are logged)
func WithVerbose(on bool) Option {
	return func(h *QueryHook) {
		h.verbose = on
	}
}

// FromEnv configures the hook using the environment variable value.
// For example, WithEnv("BUNDEBUG"):
//   - BUNDEBUG=0 - disables the hook.
//   - BUNDEBUG=1 - enables the hook.
//   - BUNDEBUG=2 - enables the hook and verbose mode.
func FromEnv(keys ...string) Option {
	if len(keys) == 0 {
		keys = []string{"BUNDEBUG"}
	}
	return func(h *QueryHook) {
		for _, key := range keys {
			if env, ok := os.LookupEnv(key); ok {
				h.enabled = env != "" && env != "0"
				h.verbose = env == "2"
				break
			}
		}
	}
}

func WithErrorLevel(level slog.Level) Option {
	return func(h *QueryHook) {
		h.errorLevel = level
	}
}

func WithQueryLevel(level slog.Level) Option {
	return func(h *QueryHook) {
		h.queryLevel = level
	}
}

func WithSlowLevel(level slog.Level) Option {
	return func(h *QueryHook) {
		h.slowLevel = level
	}
}

func WithLogSlow(duration time.Duration) Option {
	return func(h *QueryHook) {
		h.logSlow = duration
	}
}

func WithErrorTemplate(templateString string) Option {
	return func(h *QueryHook) {
		errorTemplate, err := template.New("ErrorTemplate").Parse(templateString)
		if err != nil {
			panic(err)
		}

		h.errorTemplate = errorTemplate
	}
}

func WithMessageTemplate(templateString string) Option {
	return func(h *QueryHook) {
		messageTemplate, err := template.New("MessageTemplate").Parse(templateString)
		if err != nil {
			panic(err)
		}

		h.messageTemplate = messageTemplate
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(h *QueryHook) {
		h.logger = logger
	}
}

// WithMetrics enables database metrics collection
func WithMetrics(metricsHook DatabaseMetricsHook) Option {
	return func(h *QueryHook) {
		h.metricsHook = metricsHook
		h.collectMetrics = true
	}
}

// QueryHook wraps query hook
type QueryHook struct {
	enabled bool
	verbose bool

	errorTemplate   *template.Template
	messageTemplate *template.Template

	queryLevel slog.Level
	slowLevel  slog.Level
	errorLevel slog.Level

	logSlow time.Duration

	logger *slog.Logger
	
	// metrics hook for database metrics collection
	metricsHook DatabaseMetricsHook
	collectMetrics bool
}

// DatabaseMetricsHook interface for collecting database metrics
type DatabaseMetricsHook interface {
	BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context
	AfterQuery(ctx context.Context, event *bun.QueryEvent)
}

// LogEntryVars variables made available t otemplate
type LogEntryVars struct {
	Timestamp time.Time
	Query     string
	Operation string
	Duration  time.Duration
	Error     error
}

func defaultQueryHook() *QueryHook {
	errorTemplate, err := template.New("ErrorTemplate").Parse("{{.Operation}}[{{.Duration}}]: {{.Query}}: {{.Error}}")
	if err != nil {
		panic(err)
	}
	messageTemplate, err := template.New("MessageTemplate").Parse("{{.Operation}}[{{.Duration}}]: {{.Query}}")
	if err != nil {
		panic(err)
	}

	return &QueryHook{
		enabled:         true,
		logger:          slog.Default(),
		messageTemplate: messageTemplate,
		errorTemplate:   errorTemplate,
	}
}

// NewQueryHook returns new instance
func NewQueryHook(options ...Option) *QueryHook {
	h := defaultQueryHook()

	for _, opt := range options {
		opt(h)
	}

	return h
}

// BeforeQuery calls metrics hook if enabled and does preparation
func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	// Call metrics hook if enabled
	if h.collectMetrics && h.metricsHook != nil {
		ctx = h.metricsHook.BeforeQuery(ctx, event)
	}
	return ctx
}

// AfterQuery convert a bun QueryEvent into a slog message and collect metrics
func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	// Call metrics hook first (always collect metrics regardless of logging settings)
	if h.collectMetrics && h.metricsHook != nil {
		h.metricsHook.AfterQuery(ctx, event)
	}

	if !h.enabled {
		return
	}

	if !h.verbose {
		switch event.Err {
		case nil, sql.ErrNoRows, sql.ErrTxDone:
			return
		}
	}
	var level slog.Level
	var isError bool
	var msg bytes.Buffer

	now := time.Now()
	dur := now.Sub(event.StartTime)

	switch event.Err {
	case nil, sql.ErrNoRows:
		isError = false
		if h.logSlow > 0 && dur >= h.logSlow {
			level = h.slowLevel
		} else {
			level = h.queryLevel
		}
	default:
		isError = true
		level = h.errorLevel
	}

	args := &LogEntryVars{
		Timestamp: now,
		Query:     string(event.Query),
		Operation: eventOperation(event),
		Duration:  dur,
		Error:     event.Err,
	}

	if isError {
		if err := h.errorTemplate.Execute(&msg, args); err != nil {
			panic(err)
		}
	} else {
		if err := h.messageTemplate.Execute(&msg, args); err != nil {
			panic(err)
		}
	}

	switch level {
	case slog.LevelDebug:
		h.logger.DebugContext(ctx, msg.String())
	case slog.LevelInfo:
		h.logger.InfoContext(ctx, msg.String())
	case slog.LevelWarn:
		h.logger.WarnContext(ctx, msg.String())
	case slog.LevelError:
		h.logger.ErrorContext(ctx, msg.String())
	default:
		panic(fmt.Errorf("Unsupported level: %v", level))
	}

}

// taken from bun
func eventOperation(event *bun.QueryEvent) string {
	switch event.QueryAppender.(type) {
	case *bun.SelectQuery:
		return "SELECT"
	case *bun.InsertQuery:
		return "INSERT"
	case *bun.UpdateQuery:
		return "UPDATE"
	case *bun.DeleteQuery:
		return "DELETE"
	case *bun.CreateTableQuery:
		return "CREATE TABLE"
	case *bun.DropTableQuery:
		return "DROP TABLE"
	}
	return queryOperation(event.Query)
}

// taken from bun
func queryOperation(name string) string {
	if idx := strings.Index(name, " "); idx > 0 {
		name = name[:idx]
	}
	if len(name) > 16 {
		name = name[:16]
	}
	return string(name)
}
