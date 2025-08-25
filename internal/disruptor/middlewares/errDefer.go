package middlewares

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
)

var GoErrDefer = middleware.GoErrDefer(
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
