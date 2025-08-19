package metrics

import (
	"context"
)

// DiscordSessionInterface defines the interface needed for Discord metrics collection
// This prevents import cycle by not importing the disruptor package directly
type DiscordSessionInterface interface {
	GuildsLen() int
}

// DiscordCollector collects Discord-related metrics using session interface
type DiscordCollector struct {
	session DiscordSessionInterface
	metrics *SystemMetrics
}

// NewDiscordCollector creates a new Discord collector
func NewDiscordCollector(session DiscordSessionInterface) *DiscordCollector {
	return &DiscordCollector{
		session: session,
		metrics: NewSystemMetrics(),
	}
}

// CollectGuildMetrics updates guild count metrics
func (c *DiscordCollector) CollectGuildMetrics() {
	if c.session != nil {
		// Update guild count using the new metrics registry
		guildCount := float64(c.session.GuildsLen())
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
