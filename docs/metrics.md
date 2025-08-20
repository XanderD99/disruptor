# Enhanced Prometheus Metrics Documentation

This document describes the comprehensive Prometheus metrics implemented for the Disruptor Discord bot.

## Metrics Overview

All metrics use the `disruptor_` namespace prefix and are organized into subsystems for better organization.

### Database Metrics (`disruptor_database_*`)

**Query Duration Histogram**
- Name: `disruptor_database_query_duration_seconds`
- Type: Histogram
- Labels: `operation` (select, insert, update, delete, create_table, drop_table), `table`
- Buckets: Default Prometheus buckets [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
- Description: Duration of database queries in seconds

**Query Count**
- Name: `disruptor_database_queries_total`
- Type: Counter
- Labels: `operation`, `table`, `status` (success, error)
- Description: Total number of database queries

**Database Errors**
- Name: `disruptor_database_errors_total`
- Type: Counter
- Labels: `operation`, `table`, `error_type` (timeout, connection, syntax, constraint, not_found, deadlock, other)
- Description: Total number of database errors

### Scheduler Metrics (`disruptor_scheduler_*`)

**Job Duration Histogram**
- Name: `disruptor_scheduler_job_duration_seconds`
- Type: Histogram
- Labels: `handler_type` (random_voice_join), `status` (success, error)
- Buckets: Default Prometheus buckets [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
- Description: Duration of scheduler job execution in seconds

**Job Count**
- Name: `disruptor_scheduler_jobs_total`
- Type: Counter
- Labels: `handler_type`, `status`
- Description: Total number of scheduler jobs executed

**Active Jobs**
- Name: `disruptor_scheduler_active_jobs`
- Type: Gauge
- Labels: `handler_type`
- Description: Number of currently active scheduler jobs

**Queue Depth**
- Name: `disruptor_scheduler_queue_depth`
- Type: Gauge
- Description: Number of schedulers in the manager

### Voice/Audio Metrics (`disruptor_voice_*` / `disruptor_audio_*`)

**Voice Connection Attempts**
- Name: `disruptor_voice_connection_attempts_total`
- Type: Counter
- Labels: `guild_id`, `status` (success, error)
- Description: Total number of voice connection attempts

**Active Voice Connections**
- Name: `disruptor_voice_connections_active`
- Type: Gauge
- Labels: `guild_id`
- Description: Number of active voice connections

**Audio Track Events**
- Name: `disruptor_audio_track_events_total`
- Type: Counter
- Labels: `event_type` (start, end), `guild_id`
- Description: Total number of audio track events

**Audio Processing Duration**
- Name: `disruptor_audio_processing_duration_seconds`
- Type: Histogram
- Labels: `operation` (cleanup), `guild_id`
- Buckets: Default Prometheus buckets [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
- Description: Duration of audio processing operations in seconds

### Discord API Metrics (`disruptor_discord_api_*`)

**API Requests**
- Name: `disruptor_discord_api_requests_total`
- Type: Counter
- Labels: `endpoint`, `method`, `status_code`
- Description: Total number of Discord API requests

**API Request Latency**
- Name: `disruptor_discord_api_request_duration_seconds`
- Type: Histogram
- Labels: `endpoint`, `method`
- Buckets: Default Prometheus buckets [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]
- Description: Duration of Discord API requests in seconds

### System Metrics (`disruptor_system_*`)

**Goroutine Count**
- Name: `disruptor_system_goroutines`
- Type: Gauge
- Description: Number of active goroutines

**Memory Usage**
- Name: `disruptor_system_memory_bytes`
- Type: Gauge
- Labels: `type` (heap_alloc, heap_sys, heap_idle, heap_inuse, heap_released, stack_inuse, stack_sys, total_alloc, sys)
- Description: Memory usage in bytes

**Guild Count**
- Name: `disruptor_discord_guild_count`
- Type: Gauge
- Labels: `shard`
- Description: Total number of guilds the bot is in

## Integration Points

### Database Metrics
- Integrated into `pkg/slogbun` query hook
- Automatically collects metrics for all database operations
- Categorizes errors by type for better alerting

### Scheduler Metrics
- Wraps scheduler handler functions with automatic metrics collection
- Tracks job execution duration and success/failure rates
- Updates queue depth when schedulers are added/removed

### Audio/Voice Metrics
- Integrated into Lavalink track event handlers
- Collects voice connection attempt metrics in voice processing
- Tracks audio processing operations with timing

### Discord API Metrics
- Integrated into interaction middleware for slash commands
- Collects metrics for REST API calls in scheduler handlers
- Provides endpoint-level visibility into Discord API usage

### System Metrics
- Periodic collection every 30 seconds
- Monitors Go runtime metrics (goroutines, memory)
- Tracks Discord-specific metrics (guild count)

## Example Queries

### Database Performance
```promql
# Average query duration by operation
rate(disruptor_database_query_duration_seconds_sum[5m]) / rate(disruptor_database_query_duration_seconds_count[5m])

# Database error rate
rate(disruptor_database_errors_total[5m]) / rate(disruptor_database_queries_total[5m]) * 100
```

### Scheduler Performance
```promql
# Job success rate by handler type
rate(disruptor_scheduler_jobs_total{status="success"}[5m]) / rate(disruptor_scheduler_jobs_total[5m]) * 100

# 95th percentile job execution time
histogram_quantile(0.95, rate(disruptor_scheduler_job_duration_seconds_bucket[5m]))
```

### Voice/Audio Monitoring
```promql
# Voice connection success rate
rate(disruptor_voice_connection_attempts_total{status="success"}[5m]) / rate(disruptor_voice_connection_attempts_total[5m]) * 100

# Audio track events per minute
rate(disruptor_audio_track_events_total[1m]) * 60
```

### Discord API Monitoring
```promql
# API request rate by endpoint
rate(disruptor_discord_api_requests_total[5m])

# API latency 99th percentile
histogram_quantile(0.99, rate(disruptor_discord_api_request_duration_seconds_bucket[5m]))
```

### System Health
```promql
# Goroutine growth rate
rate(disruptor_system_goroutines[5m])

# Memory usage trend
disruptor_system_memory_bytes{type="heap_inuse"}
```

## Alerting Examples

### Critical Alerts
- Database error rate > 5%
- Scheduler job failure rate > 10%
- Voice connection failure rate > 20%
- API error rate > 5%

### Warning Alerts
- Database query latency p95 > 500ms
- Scheduler job latency p95 > 30s
- Audio processing latency p95 > 1s
- API latency p95 > 2s

### Capacity Alerts
- Goroutine count > 10000
- Memory usage > 1GB
- Active voice connections > 100
- Scheduler queue depth > 50