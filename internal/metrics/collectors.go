package metrics

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

type DiscordCollector struct {
	Session *discordgo.Session
}

var (
	totalGuildsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("disruptor", "discord", "guild_count"),
		"Total number of guilds the bot is in",
		[]string{"shard"},
		nil,
	)
)

func RegisterDiscordCollector(session *discordgo.Session) error {
	collector := &DiscordCollector{
		Session: session,
	}

	if err := prometheus.Register(collector); err != nil {
		if errors.As(err, &prometheus.AlreadyRegisteredError{}) {
			// If the collector is already registered, we can safely ignore this error.
			return nil
		}
		return fmt.Errorf("failed to register Discord collector: %w", err)
	}

	return nil
}

func (c *DiscordCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		totalGuildsDesc,
		prometheus.GaugeValue,
		float64(len(c.Session.State.Guilds)),
		fmt.Sprint(c.Session.ShardID),
	)
}

func (c *DiscordCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalGuildsDesc
}
