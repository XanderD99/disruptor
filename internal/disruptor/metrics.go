package disruptor

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalGuildsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("disruptor", "discord", "guild_count"),
		"Total number of guilds the bot is in",
		[]string{"shard"},
		nil,
	)
)

func (s *Session) RegisterCollector() error {
	if err := prometheus.Register(s); err != nil {
		if errors.As(err, &prometheus.AlreadyRegisteredError{}) {
			return nil
		}
		return err
	}

	return nil
}

func (s *Session) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		totalGuildsDesc,
		prometheus.GaugeValue,
		float64(s.Caches().GuildsLen()),
	)
}

func (s *Session) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalGuildsDesc
}
