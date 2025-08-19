package disruptor

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"

	"github.com/XanderD99/disruptor/pkg/logging"
)

var loggerMiddleware handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(event *handler.InteractionEvent) error {
		logger := event.Client().Logger().With(
			slog.Group("interaction", slog.Any("id", event.Interaction.ID())),
			slog.Group("channel", slog.Any("id", event.Channel().ID())),
			slog.Group("guild", slog.Any("id", event.GuildID())),
		)

		// Add logger to context
		event.Ctx = logging.AddToContext(event.Ctx, logger)

		logger.DebugContext(event.Ctx, "handling interaction", slog.Any("interaction", event.Interaction), slog.Any("variables", event.Vars))

		if err := next(event); err != nil {
			logger.ErrorContext(event.Ctx, "error handling interaction", slog.Any("error", err))
			return err
		}

		logger.InfoContext(event.Ctx, "interaction handled successfully")
		return nil
	}
}

var errDeferMiddleware = middleware.GoErrDefer(
	func(e *handler.InteractionEvent, err error) {
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Embeds: &[]discord.Embed{
				{
					Title:       "⚠️ Command failed",
					Description: err.Error(),
					Color:       0xFF0000, // Red color for error
				},
			},
		})
		if err != nil {
			e.Client().Logger().Error("Failed to update interaction response", slog.Any("error", err))
		}
	},
	discord.InteractionTypeApplicationCommand,
	false,
	false,
)
