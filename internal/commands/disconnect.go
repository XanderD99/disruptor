package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/pkg/util"
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
	client := event.Client()
	guildID := event.GuildID()

	if err := client.UpdateVoiceState(event.Ctx, *guildID, nil, false, true); err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

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
