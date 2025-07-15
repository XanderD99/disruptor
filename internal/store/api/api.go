package api

import (
	"context"
	"time"

	"resty.dev/v3"

	"github.com/XanderD99/discord-disruptor/internal/store"
)

// Config holds the configuration for the Strapi store
type Config struct {
	// 🌐 The base URL for the Strapi API
	BaseURL string `env:"BASE_URL" default:"http://localhost:1337/api"`

	// 🔐 Authentication scheme for the Strapi
	AuthScheme string `env:"AUTH_SCHEME" default:"Bearer"`
	// 🔑 Authentication token for accessing the Strapi API
	AuthToken string `env:"AUTH_TOKEN"`

	// 🛠️ Enable debug logging for Strapi API requests
	Debug bool `env:"DEBUG" default:"false"`

	// 🔁 Number of retries for failed requests
	RetryCount int `env:"RETRY_COUNT" default:"3"`
	// ⏳ Time to wait between retries
	RetryWaitTime time.Duration `env:"RETRY_WAIT_TIME" default:"1s"`
	// ⏳ Maximum time to wait for retries
	RetryMaxWaitTime time.Duration `env:"RETRY_MAX_WAIT_TIME" default:"5s"`
}

// Store is an implementation of the Store interface using Strapi
type Store struct {
	config Config

	client *resty.Client

	guilds   store.GuildStore
	sounds   store.SoundStore
	channels store.ChannelStore
}

type Response[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta,omitempty"` // Optional metadata field
}

type Meta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	PageCount int `json:"pageCount"`
	Total     int `json:"total"`
}

// New creates a new Strapi store with the given configuration
func New(config Config) store.Store {
	client := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBaseURL(config.BaseURL).
		SetAuthScheme(config.AuthScheme).
		SetAuthToken(config.AuthToken).
		SetDebug(config.Debug).
		SetRetryCount(config.RetryCount).
		SetRetryWaitTime(config.RetryWaitTime).
		SetRetryMaxWaitTime(config.RetryMaxWaitTime)

	return &Store{
		client: client,
		config: config,
		guilds: newGuildStore(client),
		sounds: newSoundStore(client),
	}
}

// Guilds returns the guild store
func (s *Store) Guilds() store.GuildStore {
	return s.guilds
}

// Sounds returns the sound store
func (s *Store) Sounds() store.SoundStore {
	return s.sounds
}

// Channels returns the channel store
func (s *Store) Channels() store.ChannelStore {
	return s.channels
}

// Open initializes the store
func (s *Store) Open(_ context.Context) error {
	// TODO: Implement initialization logic for Strapi
	return nil
}

// Close closes the store
func (s *Store) Close() error {
	return s.client.Close()
}
