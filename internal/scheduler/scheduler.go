package scheduler

import (
	"time"
)

type Scheduler interface {
	AddGuild(guildID string) error
	RemoveGuild(guildID string) error
	GetGuilds() []string
	GetInterval() time.Duration
	GetNextIntervalTime() time.Time
	Start()
	Stop() error
	IsRunning() bool
	UpdateOptions(opts ...Option[scheduler]) error
}

var _ Scheduler = (*scheduler)(nil)

type Option[T any] func(*T)

func WithInterval(interval time.Duration) Option[scheduler] {
	return func(ig *scheduler) {
		ig.interval = interval
	}
}
