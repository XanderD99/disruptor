package logging

import (
	"context"
	"log/slog"
	"os"

	slogdiscord "github.com/betrayy/slog-discord"
	"github.com/grafana/loki-client-go/loki"
	"github.com/lmittmann/tint"
	slogloki "github.com/samber/slog-loki/v3"
	slogmulti "github.com/samber/slog-multi"
)

type Config struct {
	// 📜 Log level for the bot (e.g., debug, info, warn, error)
	Level slog.Level `env:"LEVEL" default:"debug"`
	// ✨ Enable pretty-printed logs for human readability
	PrettyPrint bool `env:"PRETTY" default:"true"`
	// 🌈 Add colors to logs for better visibility
	Colors bool `env:"COLORS" default:"true"`
	// 🗂️ Include short file paths in log messages for debugging
	AddSource bool `env:"SOURCE" default:"false"`

	Discord struct {
		// 📡 Discord webhook URL for sending log messages
		Webhook string `env:"WEBHOOK"`

		// 📉 Minimum log level for Discord messages, defaults to warn level
		MinLevel slog.Level `env:"MIN_LEVEL" default:"warn"`

		// 📦 Whether to wait for Discord messages to be sent before continuing
		Sync bool `env:"SYNC" default:"false"`
	} `envPrefix:"DISCORD_"`

	Loki struct {
		// 📡 Loki endpoint URL for sending log messages
		Endpoint string `env:"ENDPOINT"`

		// 📉 Minimum log level for Loki messages, defaults to info level
		MinLevel slog.Level `env:"MIN_LEVEL" default:"debug"`

		// 🏷️ Labels to attach to Loki log entries, in key=value format
		Labels map[string]string `env:"LABELS" envSeparator:","`

		// 📦 Whether to wait for Loki messages to be sent before continuing
		Sync bool `env:"SYNC" default:"false"`
	} `envPrefix:"LOKI_"`
}

func New(cfg Config) (*slog.Logger, error) {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	handlers := make([]slog.Handler, 0)

	if cfg.Discord.Webhook != "" {
		discordHandler, err := slogdiscord.NewDiscordHandler(
			cfg.Discord.Webhook,
			slogdiscord.WithMinLevel(cfg.Discord.MinLevel),
			slogdiscord.WithSyncMode(cfg.Discord.Sync),
		)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, discordHandler)
	}

	if cfg.Loki.Endpoint != "" {
		config, err := loki.NewDefaultConfig(cfg.Loki.Endpoint)
		if err != nil {
			return nil, err
		}
		client, err := loki.New(config)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, slogloki.Option{Level: cfg.Loki.MinLevel, Client: client}.NewLokiHandler())
	}

	if cfg.PrettyPrint {
		handlers = append(handlers, tint.NewHandler(os.Stdout, &tint.Options{
			AddSource: cfg.AddSource,
			Level:     cfg.Level,
		}))
	} else {
		handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))
	}

	logger := slog.New(slogmulti.Fanout(handlers...))
	slog.SetDefault(logger) // set global logger

	return logger, nil
}

type contextKey string

const loggerKey contextKey = "logger"

func GetFromContext(ctx context.Context) *slog.Logger {
	logger := ctx.Value(loggerKey)
	if logger == nil {
		return slog.Default()
	}

	if l, ok := logger.(*slog.Logger); ok {
		return l
	}

	// Fallback to default logger if type assertion fails
	return slog.Default()
}

func AddToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
