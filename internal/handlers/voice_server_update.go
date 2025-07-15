package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/events"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"

	"github.com/XanderD99/discord-disruptor/internal/lavalink"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

func VoiceServerUpdate(logger *slog.Logger, lava lavalink.Lavalink, store store.Store) func(*events.VoiceServerUpdate) {
	logger = logger.With(slog.String("event", "voice_server_update"))

	return func(vsu *events.VoiceServerUpdate) {
		logger = logger.With(slog.String("guild.id", vsu.GuildID.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		lava.OnVoiceServerUpdate(ctx, vsu.GuildID, vsu.Token, *vsu.Endpoint)

		sound, err := store.Sounds().Random(ctx)
		if err != nil {
			logger.Error("Failed to get random sound", slog.Any("error", err))
			return
		}

		results, err := lava.BestNode().LoadTracks(ctx, sound.URL)
		if err != nil {
			logger.Error("Failed to load tracks", slog.Any("error", err), slog.String("sound_url", sound.URL))
			return
		}

		track, ok := results.Data.(disgolavalink.Track)
		if !ok {
			logger.Error("Loaded track is not of type disgolavalink.Track", slog.Any("data_type", fmt.Sprintf("%T", results.Data)))
			return
		}

		player := lava.Player(vsu.GuildID)
		if err := player.Update(ctx, disgolavalink.WithTrack(track)); err != nil {
			logger.Error("Failed to update player with new track", slog.Any("error", err))
			return
		}
	}
}
