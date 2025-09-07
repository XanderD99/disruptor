package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/util"
	"github.com/XanderD99/disruptor/pkg/logging"
)

type play struct{}

func Play() disruptor.Command {
	return play{}
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
	// Get logger from context (added by the middleware)
	logger := logging.GetFromContext(event.Ctx)

	client := event.Client()

	voiceState, ok := client.Caches.VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return fmt.Errorf("you need to be in a voice channel to use this command")
	}

	if client.Caches.GuildSoundboardSoundsLen(*event.GuildID()) == 0 {
		return fmt.Errorf("there are no soundboard sounds available")
	}

	me, ok := client.Caches.Member(*event.GuildID(), event.Client().ID())
	if !ok {
		return fmt.Errorf("could not find myself in guild cache")
	}

	channel, ok := client.Caches.Channel(*voiceState.ChannelID)
	if !ok {
		return fmt.Errorf("could not find voice channel in cache")
	}

	permissions := client.Caches.MemberPermissionsInChannel(channel, me)
	if !util.HasVoicePermissions(permissions) {
		return fmt.Errorf("I need the `SPEAK`, `CONNECT`, and `VIEW_CHANNEL` permissions to use this command")
	}

	logger.DebugContext(event.Ctx, "user in voice channel", "channel.id", voiceState.ChannelID)

	sound, err := util.GetRandomSound(client, *event.GuildID())
	if err != nil {
		return fmt.Errorf("failed to get random sound: %w", err)
	}

	content := fmt.Sprintf("Playing %s in <#%s>", sound.Name, voiceState.ChannelID.String())
	response := discord.NewMessageUpdateBuilder().SetContent(content).Build()

	if _, err := event.UpdateInteractionResponse(response); err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}
	go func() { // fire and forget. If we don't do that here the sound could play longer than the max amount of time that discord allows between interaction and response
		if err := util.PlaySound(event.Ctx, client, *event.GuildID(), *voiceState.ChannelID, sound.URL()); err != nil {
			logger.ErrorContext(event.Ctx, "failed to play sound", "error", err)
		}
	}()

	return nil
}

var _ disruptor.Command = (*play)(nil)
