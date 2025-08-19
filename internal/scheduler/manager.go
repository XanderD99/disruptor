package scheduler

import (
	"log/slog"
	"sync"
	"time"

	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
)

type Manager interface {
	Start() error
	Stop() error

	AddScheduler(opts ...Option[scheduler]) error
	GetScheduler(interval time.Duration) (Scheduler, bool)
}

type manager struct {
	schedulers map[time.Duration]Scheduler

	// Dependencies
	session  *disruptor.Session
	db       *bun.DB
	lavalink lavalink.Lavalink
	logger   *slog.Logger

	mu sync.RWMutex
}
