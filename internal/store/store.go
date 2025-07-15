package store

import (
	"context"
	"time"
)

type Guild struct {
	ID string

	Settings GuildSettings
}

type GuildSettings struct {
	Interval time.Duration
	Chance   float64
	Enabled  bool
}

func DefaultGuildSettings() GuildSettings {
	return GuildSettings{
		Interval: time.Hour * 2, // Default interval of 30 seconds
		Chance:   0.5,           // Default chance of 50%
		Enabled:  true,          // Enabled by default
	}
}

type Channel struct {
	ID       string
	GuildID  string
	Settings ChannelSettings
}

type ChannelSettings struct {
	Enabled bool
}

func DefaultChannelSettings() ChannelSettings {
	return ChannelSettings{
		Enabled: true, // Enabled by default
	}
}

type Sound struct {
	ID   string
	Name string
	URL  string
}

type FindOptions struct {
	Sort       any
	Pagination *Pagination
	Filter     *Filter
}

type Pagination struct {
	Limit  int
	Offset int
}

func DefaultFindOptions() *FindOptions {
	return &FindOptions{
		Pagination: nil,
		Sort:       nil, // No default sorting
		Filter:     nil, // No default filtering
	}
}

func (a *FindOptions) Apply(opts ...FindOption) {
	for _, opt := range opts {
		opt(a)
	}
}

type FindOption func(options *FindOptions)

func WithPagination(limit, offset int) FindOption {
	return func(options *FindOptions) {
		options.Pagination = &Pagination{
			Limit:  limit,
			Offset: offset,
		}
	}
}

func WithSort(sort any) FindOption {
	return func(options *FindOptions) {
		options.Sort = sort
	}
}

func WithFilter(filter *Filter) FindOption {
	return func(options *FindOptions) {
		options.Filter = filter
	}
}

type CrudStore[T any] interface {
	Create(ctx context.Context, item T) (T, error)
	Update(ctx context.Context, id string, item T) error
	Delete(ctx context.Context, id string) error
	Find(ctx context.Context, opts ...FindOption) ([]T, int, error)
	FindByID(ctx context.Context, id string) (T, error)
}

type GuildStore interface {
	CrudStore[Guild]
}

type ChannelStore interface {
	CrudStore[Channel]
}

type SoundStore interface {
	Find(ctx context.Context, opts ...FindOption) ([]Sound, error)
	FindByID(ctx context.Context, id string) (Sound, error)
	Random(ctx context.Context) (Sound, error)
}

// Store defines the interface for storing and retrieving server and channel settings
type Store interface {
	Guilds() GuildStore
	Channels() ChannelStore
	Sounds() SoundStore

	// Utility methods
	Close() error
	Open(ctx context.Context) error
}
