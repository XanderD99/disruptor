package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/lavalink"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/pkg/db"
	"github.com/XanderD99/disruptor/pkg/util"
)

type Handler interface {
	handle(ctx context.Context, interval time.Duration) error
}

type handler struct {
	session  *disruptor.Session
	db       db.Database
	lavalink lavalink.Lavalink // Assuming lavalink is an interface defined in your project

	workerPool chan struct{} // Semaphore for controlling concurrent workers
}

var maxWorkers = 10 // Maximum number of concurrent workers

func NewHandler(session *disruptor.Session, db db.Database, lavalink lavalink.Lavalink) Handler {
	return handler{
		session:    session,
		db:         db,
		lavalink:   lavalink,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

func (h handler) handle(ctx context.Context, interval time.Duration) error {
	// Create a context with timeout for this batch
	guilds, err := h.getEligibleGuilds(ctx, interval)
	if err != nil {
		return fmt.Errorf("failed to find guilds: %w", err)
	}

	// Process guilds with worker pool
	return h.processGuildsWithPool(ctx, guilds)
}

func (h handler) processGuildsWithPool(ctx context.Context, guilds []models.Guild) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(guilds))

	process := func(g models.Guild) {
		defer func() {
			<-h.workerPool // Release worker
			wg.Done()
		}()

		if err := h.processGuild(ctx, g); err != nil {
			select {
			case errChan <- err:
			default: // Don't block if channel is full
			}
		}
	}

	for _, guild := range guilds {
		wg.Add(1)

		// Acquire worker from pool
		select {
		case h.workerPool <- struct{}{}:
			go process(guild)
		case <-ctx.Done():
			wg.Done()
			return ctx.Err()
		}
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors (you might want to log them instead of returning)
	for err := range errChan {
		h.session.Logger().Error("Failed to process guild", "error", err)
	}

	return nil // Don't fail the entire batch for individual guild failures
}

func (h handler) processGuild(ctx context.Context, guild models.Guild) error {
	// guildID := snowflake.MustParse(guild.ID)

	if _, ok := h.session.Caches().Guild(guild.ID); !ok {
		h.session.Logger().Warn("Guild not found in cache, skipping", "guild_id", guild.ID)
		return nil // Skip if guild is not in cache
	}

	// Get available voice channels
	channels, err := h.getAvailableVoiceChannels(ctx, guild.ID)
	if err != nil {
		return fmt.Errorf("failed to get channels for guild %s: %w", guild.ID, err)
	}

	if len(channels) == 0 {
		return nil // No available channels, skip
	}

	// Select a random channel
	channel := channels[util.RandomInt(0, len(channels)-1)]

	channelID := channel.ID()
	if err := h.session.UpdateVoiceState(ctx, guild.ID, &channelID, false, true); err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	return nil
}

func (h handler) getAvailableVoiceChannels(ctx context.Context, guildID snowflake.ID) ([]discord.GuildChannel, error) {
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

		audioChannel, ok := h.session.Caches().GuildAudioChannel(channel.ID())
		if !ok {
			continue
		}
		members := h.session.Caches().AudioChannelMembers(audioChannel)
		if len(members) == 0 {
			continue
		}

		permissions := h.session.Caches().MemberPermissionsInChannel(audioChannel, member)
		if !permissions.Has(discord.PermissionViewChannel) || !permissions.Has(discord.PermissionConnect) {
			continue
		}
		filtered = append(filtered, channel)
	}

	return filtered, nil
}

func (h handler) getEligibleGuilds(ctx context.Context, interval time.Duration) ([]models.Guild, error) {
	chance := util.RandomFloat(0, 100) // Use float for better precision

	logger := h.session.Logger().With(
		slog.Duration("interval", interval),
		slog.Float64("chance", chance),
	)

	filter := map[string]any{
		"interval": int64(interval),
		"chance": map[string]any{
			"$gte": chance,
		},
	}
	guilds := make([]models.Guild, 0)
	if err := h.db.Find(ctx, &guilds, db.WithFilters(filter)); err != nil {
		return nil, fmt.Errorf("failed to find guilds: %w", err)
	}

	count := len(guilds)
	if count == 0 {
		logger.Info("No eligible guilds found for processing")
		return nil, nil // No eligible guilds
	}

	h.session.Logger().Info("Fetched all eligible guilds", slog.Int("count", count))

	return guilds, nil
}
