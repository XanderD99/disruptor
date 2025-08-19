package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Registry holds all metrics for the application
type Registry struct {
	// Database metrics
	DatabaseQueryDuration *prometheus.HistogramVec
	DatabaseQueryTotal    *prometheus.CounterVec
	DatabaseErrors        *prometheus.CounterVec

	// Scheduler metrics
	SchedulerJobDuration    *prometheus.HistogramVec
	SchedulerJobTotal       *prometheus.CounterVec
	SchedulerActiveJobs     *prometheus.GaugeVec
	SchedulerQueueDepth     prometheus.Gauge

	// Audio/Voice metrics
	VoiceConnectionAttempts *prometheus.CounterVec
	VoiceConnections        *prometheus.GaugeVec
	AudioTrackEvents        *prometheus.CounterVec
	AudioProcessingDuration *prometheus.HistogramVec

	// Discord API metrics
	DiscordAPIRequests *prometheus.CounterVec
	DiscordAPILatency  *prometheus.HistogramVec

	// System metrics
	GoroutineCount prometheus.Gauge
	MemoryUsage    *prometheus.GaugeVec

	// Existing metrics (keeping for compatibility)
	TotalGuilds *prometheus.GaugeVec
}

var (
	registry *Registry
	once     sync.Once
)

// GetRegistry returns the singleton metrics registry
func GetRegistry() *Registry {
	once.Do(func() {
		registry = newRegistry()
	})
	return registry
}

func newRegistry() *Registry {
	r := &Registry{
		// Database metrics
		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "disruptor",
				Subsystem: "database",
				Name:      "query_duration_seconds",
				Help:      "Duration of database queries in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"operation", "table"},
		),

		DatabaseQueryTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "disruptor",
				Subsystem: "database",
				Name:      "queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		DatabaseErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "disruptor",
				Subsystem: "database",
				Name:      "errors_total",
				Help:      "Total number of database errors",
			},
			[]string{"operation", "table", "error_type"},
		),

		// Scheduler metrics
		SchedulerJobDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "disruptor",
				Subsystem: "scheduler",
				Name:      "job_duration_seconds",
				Help:      "Duration of scheduler job execution in seconds",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
			},
			[]string{"handler_type", "status"},
		),

		SchedulerJobTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "disruptor",
				Subsystem: "scheduler",
				Name:      "jobs_total",
				Help:      "Total number of scheduler jobs executed",
			},
			[]string{"handler_type", "status"},
		),

		SchedulerActiveJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "disruptor",
				Subsystem: "scheduler",
				Name:      "active_jobs",
				Help:      "Number of currently active scheduler jobs",
			},
			[]string{"handler_type"},
		),

		SchedulerQueueDepth: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "disruptor",
				Subsystem: "scheduler",
				Name:      "queue_depth",
				Help:      "Number of schedulers in the manager",
			},
		),

		// Audio/Voice metrics
		VoiceConnectionAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "disruptor",
				Subsystem: "voice",
				Name:      "connection_attempts_total",
				Help:      "Total number of voice connection attempts",
			},
			[]string{"guild_id", "status"},
		),

		VoiceConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "disruptor",
				Subsystem: "voice",
				Name:      "connections_active",
				Help:      "Number of active voice connections",
			},
			[]string{"guild_id"},
		),

		AudioTrackEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "disruptor",
				Subsystem: "audio",
				Name:      "track_events_total",
				Help:      "Total number of audio track events",
			},
			[]string{"event_type", "guild_id"},
		),

		AudioProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "disruptor",
				Subsystem: "audio",
				Name:      "processing_duration_seconds",
				Help:      "Duration of audio processing operations in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2, 5},
			},
			[]string{"operation", "guild_id"},
		),

		// Discord API metrics
		DiscordAPIRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "disruptor",
				Subsystem: "discord_api",
				Name:      "requests_total",
				Help:      "Total number of Discord API requests",
			},
			[]string{"endpoint", "method", "status_code"},
		),

		DiscordAPILatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "disruptor",
				Subsystem: "discord_api",
				Name:      "request_duration_seconds",
				Help:      "Duration of Discord API requests in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2, 5},
			},
			[]string{"endpoint", "method"},
		),

		// System metrics
		GoroutineCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "disruptor",
				Subsystem: "system",
				Name:      "goroutines",
				Help:      "Number of active goroutines",
			},
		),

		MemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "disruptor",
				Subsystem: "system",
				Name:      "memory_bytes",
				Help:      "Memory usage in bytes",
			},
			[]string{"type"},
		),

		// Existing metrics for compatibility
		TotalGuilds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "disruptor",
				Subsystem: "discord",
				Name:      "guild_count",
				Help:      "Total number of guilds the bot is in",
			},
			[]string{"shard"},
		),
	}

	return r
}

// MustRegister registers all metrics with Prometheus, panicking on error
func (r *Registry) MustRegister() {
	prometheus.MustRegister(
		r.DatabaseQueryDuration,
		r.DatabaseQueryTotal,
		r.DatabaseErrors,
		r.SchedulerJobDuration,
		r.SchedulerJobTotal,
		r.SchedulerActiveJobs,
		r.SchedulerQueueDepth,
		r.VoiceConnectionAttempts,
		r.VoiceConnections,
		r.AudioTrackEvents,
		r.AudioProcessingDuration,
		r.DiscordAPIRequests,
		r.DiscordAPILatency,
		r.GoroutineCount,
		r.MemoryUsage,
		r.TotalGuilds,
	)
}

// Register registers all metrics with Prometheus, returning any error
func (r *Registry) Register() error {
	collectors := []prometheus.Collector{
		r.DatabaseQueryDuration,
		r.DatabaseQueryTotal,
		r.DatabaseErrors,
		r.SchedulerJobDuration,
		r.SchedulerJobTotal,
		r.SchedulerActiveJobs,
		r.SchedulerQueueDepth,
		r.VoiceConnectionAttempts,
		r.VoiceConnections,
		r.AudioTrackEvents,
		r.AudioProcessingDuration,
		r.DiscordAPIRequests,
		r.DiscordAPILatency,
		r.GoroutineCount,
		r.MemoryUsage,
		r.TotalGuilds,
	}

	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			return err
		}
	}

	return nil
}