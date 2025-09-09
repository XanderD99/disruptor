package middlewares

import (
	"log/slog"

	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/oteldisgo"

	"github.com/XanderD99/disruptor/pkg/logging"
)

var Otel = oteldisgo.Middleware("disruptor")

var Logger handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(event *handler.InteractionEvent) error {
		logger := event.Client().Logger.With(
			slog.Group("interaction", slog.Any("id", event.Interaction.ID())),
			slog.Group("channel", slog.Any("id", event.Channel().ID())),
			slog.Group("guild", slog.Any("id", event.GuildID())),
		)

		// Add logger to context
		event.Ctx = logging.AddToContext(event.Ctx, logger)

		logger.DebugContext(event.Ctx, "handling interaction")
		if err := next(event); err != nil {
			logger.ErrorContext(event.Ctx, "interaction failed", slog.Any("error", err))
			return err
		}
		return nil
	}
}
