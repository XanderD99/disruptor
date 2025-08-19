package listeners

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"

	"github.com/XanderD99/disruptor/internal/lavalink"
)

func VoiceStateUpdate(logger *slog.Logger, lava lavalink.Lavalink) func(*events.GuildVoiceStateUpdate) {
	logger = logger.With(slog.String("event", "voice_state_update"))

	return func(vsu *events.GuildVoiceStateUpdate) {
		// filter all non bot voice state updates out
		if vsu.VoiceState.UserID != vsu.Client().ID() {
			logger.Debug("Ignoring voice state update for non-bot user", slog.String("user.id", vsu.VoiceState.UserID.String()))
			return
		}

		lava.OnVoiceStateUpdate(context.Background(), vsu.VoiceState.GuildID, vsu.VoiceState.ChannelID, vsu.VoiceState.SessionID)
	}
}
