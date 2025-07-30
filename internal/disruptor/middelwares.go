package disruptor

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
)

var loggerMiddleware handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(event *handler.InteractionEvent) error {
		logger := event.Client().Logger().With(
			slog.Any("interaction.id", event.Interaction.ID()),
			slog.Any("channel.id", event.Channel().ID()),
			slog.Any("guild.id", event.GuildID()),
		)

		logger.InfoContext(event.Ctx, "handling interaction", slog.Any("interaction", event.Interaction), slog.Any("vars", event.Vars))

		if err := next(event); err != nil {
			logger.ErrorContext(event.Ctx, "error handling interaction", slog.Any("error", err))
			return err
		}
		return nil
	}
}

var errDeferMiddleware = middleware.GoErrDefer(
	func(e *handler.InteractionEvent, err error) {
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Embeds: &[]discord.Embed{
				{
					Title:       "Error",
					Description: fmt.Sprintf("An error occurred while processing your request: %s", err.Error()),
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
