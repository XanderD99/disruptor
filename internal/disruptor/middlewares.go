package disruptor

import (
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"

	"github.com/XanderD99/disruptor/internal/metrics"
	"github.com/XanderD99/disruptor/pkg/logging"
)

var loggerMiddleware handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(event *handler.InteractionEvent) error {
		// Start timer for interaction handling
		discordMetrics := metrics.NewDiscordAPIMetrics()
		startTime := time.Now()
		
		logger := event.Client().Logger().With(
			slog.Group("interaction", slog.Any("id", event.Interaction.ID())),
			slog.Group("channel", slog.Any("id", event.Channel().ID())),
			slog.Group("guild", slog.Any("id", event.GuildID())),
		)

		// Add logger to context
		event.Ctx = logging.AddToContext(event.Ctx, logger)

		logger.DebugContext(event.Ctx, "handling interaction", slog.Any("interaction", event.Interaction), slog.Any("variables", event.Vars))

		err := next(event)
		
		// Record interaction metrics
		duration := time.Since(startTime)
		statusCode := 200 // Assume success
		if err != nil {
			statusCode = 500
			logger.ErrorContext(event.Ctx, "error handling interaction", slog.Any("error", err))
		} else {
			logger.InfoContext(event.Ctx, "interaction handled successfully")
		}
		
		// Record metrics based on interaction type
		endpoint := "interaction"
		if event.Interaction.Type() == discord.InteractionTypeApplicationCommand {
			endpoint = "slash_command"
		}
		discordMetrics.RecordAPIRequest(endpoint, "POST", statusCode, duration)
		
		return err
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
