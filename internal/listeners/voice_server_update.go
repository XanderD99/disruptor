package listeners

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"

	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/pkg/util"
)

func VoiceServerUpdate(logger *slog.Logger, lava lavalink.Lavalink) func(*events.VoiceServerUpdate) {
	logger = logger.With(slog.String("event", "voice_server_update"))

	return func(vsu *events.VoiceServerUpdate) {
		logger = logger.With(slog.String("guild.id", vsu.GuildID.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		lava.OnVoiceServerUpdate(ctx, vsu.GuildID, vsu.Token, *vsu.Endpoint)

		client := vsu.Client()

		sounds := make([]discord.SoundboardSound, 0)
		client.Caches().GuildSoundboardSoundsForEach(vsu.GuildID, func(soundboardSound discord.SoundboardSound) {
			sounds = append(sounds, soundboardSound)
		})

		// pick random sound from the list
		index := util.RandomInt(0, len(sounds)-1)
		sound := sounds[index]

		results, err := lava.BestNode().LoadTracks(ctx, sound.URL())
		if err != nil {
			logger.Error("Failed to load tracks", slog.Any("error", err), slog.String("sound.url", sound.URL()))
			return
		}

		track, ok := results.Data.(disgolavalink.Track)
		if !ok {
			logger.Error("Loaded track is not of type disgolavalink.Track", slog.Any("data.type", fmt.Sprintf("%T", results.Data)))
			return
		}

		player := lava.Player(vsu.GuildID)
		if err := player.Update(ctx, disgolavalink.WithTrack(track)); err != nil {
			logger.Error("Failed to update player with new track", slog.Any("error", err))
			return
		}
	}
}
