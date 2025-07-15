package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
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
}

func New(cfg Config) (*slog.Logger, error) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	if cfg.PrettyPrint {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource: cfg.AddSource,
			Level:     cfg.Level,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger) // set global logger

	return logger, nil
}

type contextKey string

const loggerKey contextKey = "logger"

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
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

func PutLoggerInContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
