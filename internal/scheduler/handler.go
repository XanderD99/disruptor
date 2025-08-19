package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/pkg/util"
)

type Handler interface {
	handle(ctx context.Context, interval time.Duration) error
}

type handler struct {
	session  *disruptor.Session
	db       *bun.DB
	lavalink lavalink.Lavalink // Assuming lavalink is an interface defined in your project

}

func NewHandler(session *disruptor.Session, db *bun.DB, lavalink lavalink.Lavalink) Handler {
	return handler{
		session:  session,
		db:       db,
		lavalink: lavalink,
	}
}

func (h handler) handle(ctx context.Context, interval time.Duration) error {
	chance := util.RandomFloat(0, 100) // Use float for better precision

	// Create a context with timeout for this batch
	guilds, err := h.getEligibleGuilds(ctx, chance, interval)
	if err != nil {
		return fmt.Errorf("failed to find guilds: %w", err)
	}

	if len(guilds) == 0 {
		h.session.Logger().Info("No eligible guilds found for interval", slog.Duration("interval", interval))
		return nil
	}

	// Process guilds with worker pool
	return h.processGuildsWithPool(ctx, guilds)
}

func (h handler) processGuildsWithPool(ctx context.Context, guilds []models.Guild) error {
	maxWorkers := int(math.Max(1, math.Sqrt(float64(len(guilds)))))
	return util.ProcessWithWorkerPool(ctx, guilds, maxWorkers, func(ctx context.Context, guild models.Guild) {
		if err := h.processGuild(ctx, guild); err != nil {
			h.session.Logger().Error("Failed to process guild", slog.Any("guild.id", guild.Snowflake), slog.Any("error", err))
		}
	})
}

func (h handler) processGuild(ctx context.Context, guild models.Guild) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, ok := h.session.Caches().Guild(guild.Snowflake); !ok {
		h.session.Logger().Warn("Guild not found in cache, skipping", slog.Any("guild.id", guild.Snowflake))
		return nil // Skip if guild is not in cache
	}

	// Get available voice channels
	channels, err := h.getAvailableVoiceChannels(ctx, guild.Snowflake)
	if err != nil {
		return fmt.Errorf("failed to get channels for guild %s: %w", guild.Snowflake, err)
	}

	if len(channels) == 0 {
		return nil // No available channels, skip
	}

	// Select a random channel
	channelID := randomChannelID(channels)
	if err := h.session.UpdateVoiceState(ctx, guild.Snowflake, &channelID, false, true); err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	return nil
}

func randomChannelID(channels []discord.GuildChannel) snowflake.ID {
	return channels[util.RandomInt(0, len(channels)-1)].ID()
}

func (h handler) getAvailableVoiceChannels(ctx context.Context, guildID snowflake.ID) ([]discord.GuildChannel, error) {
	if h.session.Caches().GuildSoundboardSoundsLen(guildID) == 0 {
		return nil, fmt.Errorf("there are no soundboard sounds available")
	}

	channels, err := h.session.Rest().GetGuildChannels(guildID, rest.WithCtx(ctx))
	if err != nil {
		return nil, err
	}

	member, ok := h.session.Caches().Member(guildID, h.session.ID())
	if !ok {
		return nil, fmt.Errorf("bot is not a member of the guild %s", guildID)
	}

	filtered := make([]discord.GuildChannel, 0, len(channels))
	for _, channel := range channels {
		if channel.Type() != discord.ChannelTypeGuildVoice {
			continue
		}

		voiceChannel, ok := h.session.Caches().GuildVoiceChannel(channel.ID())
		if !ok {
			continue
		}

		permissions := h.session.Caches().MemberPermissionsInChannel(voiceChannel, member)
		if !util.HasVoicePermissions(permissions) {
			continue
		}

		if len(h.session.Caches().AudioChannelMembers(voiceChannel)) == 0 {
			continue
		}

		filtered = append(filtered, channel)
	}

	return filtered, nil
}

func (h handler) getEligibleGuilds(ctx context.Context, chance float64, interval time.Duration) ([]models.Guild, error) {
	guilds := make([]models.Guild, 0)
	if err := h.db.NewSelect().Model(&guilds).Where("chance <= ? AND interval = ?", chance, interval).Scan(ctx, &guilds); err != nil {
		return nil, fmt.Errorf("failed to find eligible guilds: %w", err)
	}

	count := len(guilds)
	if count == 0 {
		return nil, nil // No eligible guilds
	}

	return guilds, nil
}
