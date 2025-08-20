package metrics

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

// AudioMetrics provides methods for recording audio and voice-related metrics
type AudioMetrics struct{}

// NewAudioMetrics creates a new audio metrics instance
func NewAudioMetrics() *AudioMetrics {
	return &AudioMetrics{}
}

// RecordVoiceConnectionAttempt records a voice connection attempt
func (a *AudioMetrics) RecordVoiceConnectionAttempt(guildID snowflake.ID, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	VoiceConnectionAttempts.WithLabelValues(guildID.String(), status).Inc()
}

// RecordVoiceConnectionActive records an active voice connection
func (a *AudioMetrics) RecordVoiceConnectionActive(guildID snowflake.ID) {
	VoiceConnections.WithLabelValues(guildID.String()).Inc()
}

// RecordVoiceConnectionClosed records a closed voice connection
func (a *AudioMetrics) RecordVoiceConnectionClosed(guildID snowflake.ID) {
	VoiceConnections.WithLabelValues(guildID.String()).Dec()
}

// RecordTrackEvent records audio track events (start, end, etc.)
func (a *AudioMetrics) RecordTrackEvent(eventType string, guildID snowflake.ID) {
	AudioTrackEvents.WithLabelValues(eventType, guildID.String()).Inc()
}

// RecordAudioProcessingDuration records the duration of audio processing operations
func (a *AudioMetrics) RecordAudioProcessingDuration(operation string, guildID snowflake.ID, duration time.Duration) {
	AudioProcessingDuration.WithLabelValues(operation, guildID.String()).Observe(duration.Seconds())
}

// AudioProcessingTimer provides a timer for measuring audio processing duration
type AudioProcessingTimer struct {
	metrics   *AudioMetrics
	operation string
	guildID   snowflake.ID
	startTime time.Time
}

// NewAudioProcessingTimer creates a new timer for audio processing operations
func (a *AudioMetrics) NewAudioProcessingTimer(operation string, guildID snowflake.ID) *AudioProcessingTimer {
	return &AudioProcessingTimer{
		metrics:   a,
		operation: operation,
		guildID:   guildID,
		startTime: time.Now(),
	}
}

// Finish completes the timer and records the processing duration
func (t *AudioProcessingTimer) Finish() {
	duration := time.Since(t.startTime)
	t.metrics.RecordAudioProcessingDuration(t.operation, t.guildID, duration)
}

// Voice state update metrics
func (a *AudioMetrics) RecordVoiceStateUpdate(guildID snowflake.ID, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	// Use a generic "voice_state_update" operation for tracking
	VoiceConnectionAttempts.WithLabelValues(guildID.String(), status).Inc()
}
