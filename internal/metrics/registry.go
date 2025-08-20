package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics are automatically registered with the default registry using promauto
var (
	// Database metrics
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "disruptor",
			Subsystem: "database",
			Name:      "query_duration_seconds",
			Help:      "Duration of database queries in seconds",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "disruptor",
			Subsystem: "database",
			Name:      "queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	DatabaseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "disruptor",
			Subsystem: "database",
			Name:      "errors_total",
			Help:      "Total number of database errors",
		},
		[]string{"operation", "table", "error_type"},
	)

	// Scheduler metrics
	SchedulerJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "disruptor",
			Subsystem: "scheduler",
			Name:      "job_duration_seconds",
			Help:      "Duration of scheduler job execution in seconds",
		},
		[]string{"handler_type", "status"},
	)

	SchedulerJobTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "disruptor",
			Subsystem: "scheduler",
			Name:      "jobs_total",
			Help:      "Total number of scheduler jobs executed",
		},
		[]string{"handler_type", "status"},
	)

	SchedulerActiveJobs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "disruptor",
			Subsystem: "scheduler",
			Name:      "active_jobs",
			Help:      "Number of currently active scheduler jobs",
		},
		[]string{"handler_type"},
	)

	SchedulerQueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "disruptor",
			Subsystem: "scheduler",
			Name:      "queue_depth",
			Help:      "Number of schedulers in the manager",
		},
	)

	// Audio/Voice metrics
	VoiceConnectionAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "disruptor",
			Subsystem: "voice",
			Name:      "connection_attempts_total",
			Help:      "Total number of voice connection attempts",
		},
		[]string{"guild_id", "status"},
	)

	VoiceConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "disruptor",
			Subsystem: "voice",
			Name:      "connections_active",
			Help:      "Number of active voice connections",
		},
		[]string{"guild_id"},
	)

	AudioTrackEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "disruptor",
			Subsystem: "audio",
			Name:      "track_events_total",
			Help:      "Total number of audio track events",
		},
		[]string{"event_type", "guild_id"},
	)

	AudioProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "disruptor",
			Subsystem: "audio",
			Name:      "processing_duration_seconds",
			Help:      "Duration of audio processing operations in seconds",
		},
		[]string{"operation", "guild_id"},
	)

	// Discord API metrics
	DiscordAPIRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "disruptor",
			Subsystem: "discord_api",
			Name:      "requests_total",
			Help:      "Total number of Discord API requests",
		},
		[]string{"endpoint", "method", "status_code"},
	)

	DiscordAPILatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "disruptor",
			Subsystem: "discord_api",
			Name:      "request_duration_seconds",
			Help:      "Duration of Discord API requests in seconds",
		},
		[]string{"endpoint", "method"},
	)

	// System metrics
	GoroutineCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "disruptor",
			Subsystem: "system",
			Name:      "goroutines",
			Help:      "Number of active goroutines",
		},
	)

	MemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "disruptor",
			Subsystem: "system",
			Name:      "memory_bytes",
			Help:      "Memory usage in bytes",
		},
		[]string{"type"},
	)

	// Existing metrics for compatibility
	TotalGuilds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "disruptor",
			Subsystem: "discord",
			Name:      "guild_count",
			Help:      "Total number of guilds the bot is in",
		},
		[]string{"shard"},
	)
)
