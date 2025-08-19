package metrics

import (
	"context"

	"github.com/XanderD99/disruptor/internal/disruptor"
)

// DiscordCollector collects Discord-related metrics using the new disgo session
type DiscordCollector struct {
	Session *disruptor.Session
	metrics *SystemMetrics
}

// NewDiscordCollector creates a new Discord collector
func NewDiscordCollector(session *disruptor.Session) *DiscordCollector {
	return &DiscordCollector{
		Session: session,
		metrics: NewSystemMetrics(),
	}
}

// CollectGuildMetrics updates guild count metrics
func (c *DiscordCollector) CollectGuildMetrics() {
	if c.Session != nil {
		// Update guild count using the new metrics registry
		guildCount := float64(c.Session.Caches().GuildsLen())
		c.metrics.RecordGuildCount("0", guildCount) // Using "0" as default shard ID
	}
}

// StartCollection starts periodic collection of Discord metrics
func (c *DiscordCollector) StartCollection(ctx context.Context) error {
	// Collect initial metrics
	c.CollectGuildMetrics()
	
	// Note: In a production system, you might want to collect these metrics
	// periodically, but for now we'll collect them on-demand
	return nil
}
