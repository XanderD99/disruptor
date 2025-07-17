package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
)

type play struct {
	lavalink lavalink.Lavalink
}

func Play(lavalink lavalink.Lavalink) disruptor.Command {
	return play{
		lavalink: lavalink,
	}
}

// Load implements disruptor.Command.
func (p play) Load(r handler.Router) {
	r.SlashCommand("/play", p.handle)
}

// Options implements disruptor.Command.
func (p play) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "play",
		Description: "Play a sound in your current voice channel",
	}
}

func (p play) handle(_ discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	client := event.Client()

	voiceState, ok := client.Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return fmt.Errorf("you need to be in a voice channel to use this command")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	player := p.lavalink.ExistingPlayer(*event.GuildID())
	if player != nil && player.Track() != nil {
		return fmt.Errorf("already playing in this guild. Please try again when the bot leaves the voice channel")
	}

	if err := client.UpdateVoiceState(ctx, *event.GuildID(), voiceState.ChannelID, false, true); err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	content := fmt.Sprintf("Playing in <#%s>", voiceState.ChannelID.String())
	response := discord.NewMessageUpdateBuilder().SetContent(content).Build()

	if _, err := event.UpdateInteractionResponse(response); err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*play)(nil)
