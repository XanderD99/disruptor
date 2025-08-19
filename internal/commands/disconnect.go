package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/internal/util"
	"github.com/XanderD99/disruptor/pkg/logging"
)

type disconnect struct {
	lavalink lavalink.Lavalink
}

func Disconnect(lavalink lavalink.Lavalink) disruptor.Command {
	return disconnect{
		lavalink: lavalink,
	}
}

// Load implements disruptor.Command.
func (p disconnect) Load(r handler.Router) {
	r.SlashCommand("/disconnect", p.handle)
}

// Options implements disruptor.Command.
func (p disconnect) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "disconnect",
		Description: "Disconnect the bot from the voice channel",
	}
}

func (p disconnect) handle(_ discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	// Get logger from context (added by the middleware)
	logger := logging.GetFromContext(event.Ctx)

	client := event.Client()
	guildID := event.GuildID()

	logger.DebugContext(event.Ctx, "disconnecting bot from voice channel")

	if err := client.UpdateVoiceState(event.Ctx, *guildID, nil, false, true); err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	logger.DebugContext(event.Ctx, "successfully disconnected from voice channel")

	embed := discord.NewEmbedBuilder()
	embed.SetColor(util.RGBToInteger(255, 215, 0))

	embed.SetTitle("Disconnected")
	embed.SetDescription("I have been disconnected from the voice channel.")

	msg := discord.NewMessageUpdateBuilder().SetEmbeds((embed).Build()).Build()

	if _, err := event.UpdateInteractionResponse(msg); err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*play)(nil)
