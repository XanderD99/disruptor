package scheduler

import (
	"log/slog"
	"sync"
	"time"

	"github.com/XanderD99/discord-disruptor/internal/disruptor"
	"github.com/XanderD99/discord-disruptor/internal/lavalink"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

type Manager interface {
	Start() error
	Stop() error
	GetSchedulerForGuild(guildID string) (Scheduler, error)
	AddScheduler(opts ...Option[scheduler]) error
	AddGuild(guildID string, interval time.Duration) error
	RemoveGuild(guildID string) error
}

type manager struct {
	intervalGroups map[string]Scheduler

	maxGuildsPerScheduler int

	// Dependencies
	session  *disruptor.Session
	store    store.Store
	lavalink lavalink.Lavalink
	logger   *slog.Logger

	mu sync.RWMutex
}

func WithMaxGuildsPerScheduler(maxGuilds int) Option[manager] {
	return func(m *manager) {
		m.maxGuildsPerScheduler = maxGuilds
	}
}
