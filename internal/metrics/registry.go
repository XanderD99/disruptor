package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics using OpenTelemetry for automatic registration and export
var (
	meter = otel.Meter("disruptor")

	// Database metrics - now handled by bunotel automatically
	// These are kept for backward compatibility but will use OpenTelemetry's automatic collection

	// Scheduler metrics
	SchedulerJobDuration metric.Float64Histogram
	SchedulerJobTotal    metric.Int64Counter
	SchedulerActiveJobs  metric.Int64UpDownCounter
	SchedulerQueueDepth  metric.Int64UpDownCounter

	// Audio/Voice metrics
	VoiceConnectionAttempts metric.Int64Counter
	VoiceConnections        metric.Int64UpDownCounter
	AudioTrackEvents        metric.Int64Counter
	AudioProcessingDuration metric.Float64Histogram

	// System metrics - using observability callbacks for gauge-like behavior
	// These will be implemented as async gauges

	// Existing metrics for compatibility
	TotalGuilds metric.Int64UpDownCounter
)

func init() {
	// Initialize all OpenTelemetry metrics
	var err error

	// Scheduler metrics
	SchedulerJobDuration, err = meter.Float64Histogram(
		"disruptor_scheduler_job_duration_seconds",
		metric.WithDescription("Duration of scheduler job execution in seconds"),
	)
	if err != nil {
		panic(err)
	}

	SchedulerJobTotal, err = meter.Int64Counter(
		"disruptor_scheduler_jobs_total",
		metric.WithDescription("Total number of scheduler jobs executed"),
	)
	if err != nil {
		panic(err)
	}

	SchedulerActiveJobs, err = meter.Int64UpDownCounter(
		"disruptor_scheduler_active_jobs",
		metric.WithDescription("Number of currently active scheduler jobs"),
	)
	if err != nil {
		panic(err)
	}

	SchedulerQueueDepth, err = meter.Int64UpDownCounter(
		"disruptor_scheduler_queue_depth",
		metric.WithDescription("Number of schedulers in the manager"),
	)
	if err != nil {
		panic(err)
	}

	// Audio/Voice metrics
	VoiceConnectionAttempts, err = meter.Int64Counter(
		"disruptor_voice_connection_attempts_total",
		metric.WithDescription("Total number of voice connection attempts"),
	)
	if err != nil {
		panic(err)
	}

	VoiceConnections, err = meter.Int64UpDownCounter(
		"disruptor_voice_connections_active",
		metric.WithDescription("Number of active voice connections"),
	)
	if err != nil {
		panic(err)
	}

	AudioTrackEvents, err = meter.Int64Counter(
		"disruptor_audio_track_events_total",
		metric.WithDescription("Total number of audio track events"),
	)
	if err != nil {
		panic(err)
	}

	AudioProcessingDuration, err = meter.Float64Histogram(
		"disruptor_audio_processing_duration_seconds",
		metric.WithDescription("Duration of audio processing operations in seconds"),
	)
	if err != nil {
		panic(err)
	}

	// Existing metrics for compatibility
	TotalGuilds, err = meter.Int64UpDownCounter(
		"disruptor_discord_guild_count",
		metric.WithDescription("Total number of guilds the bot is in"),
	)
	if err != nil {
		panic(err)
	}
}
