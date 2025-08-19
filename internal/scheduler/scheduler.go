package scheduler

import (
	"time"
)

type Scheduler interface {
	GetInterval() time.Duration
	GetNextIntervalTime() time.Time
	Start()
	Stop() error
	IsRunning() bool
}

type Option[T any] interface {
	Apply(*T)
	IntervalOrDefault(defaultInterval time.Duration) time.Duration
}

// Example option for interval configuration.
type intervalOption struct {
	interval time.Duration
}

func (o intervalOption) Apply(s *scheduler) {
	if s.timer != nil {
		s.timer.interval = o.interval
	}
}
func (o intervalOption) IntervalOrDefault(defaultInterval time.Duration) time.Duration {
	if o.interval > 0 {
		return o.interval
	}
	return defaultInterval
}

func WithInterval(interval time.Duration) Option[scheduler] {
	return intervalOption{interval: interval}
}
