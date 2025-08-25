package disruptor

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/XanderD99/disruptor/internal/disruptor/middlewares"
)

// Session represents the Discord bot
type Session struct {
	bot.Client
	handler.Router
}

// New creates a new Discord bot with sharding support
func New(cfg Config) (*Session, error) {
	router := handler.New()
	router.Use(
		middlewares.Otel,
		middlewares.Logger,
		middlewares.GoErrDefer,
	)
	options := []bot.ConfigOpt{}
	options = append(options, bot.WithEventListeners(router))
	options = append(options, cfg.ToSessionOpts()...)

	c, err := disgo.New(cfg.Token, options...)
	if err != nil {
		return nil, err
	}

	s := &Session{
		Client: c,
		Router: router,
	}

	return s, nil
}

func (s *Session) UpdateVoiceState(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID) error {
	return s.Client.UpdateVoiceState(ctx, guildID, channelID, true, false)
}

func (s *Session) AddCommands(commands ...Command) error {
	cmds := make([]discord.ApplicationCommandCreate, 0, len(commands))
	for _, cmd := range commands {
		if cmd == nil {
			continue // Skip nil commands
		}
		cmd.Load(s.Router)
		cmds = append(cmds, cmd.Options())
	}

	guild := snowflake.GetEnv("CONFIG_GUILDID")
	guildIDs := []snowflake.ID{}
	if guild != 0 {
		guildIDs = append(guildIDs, guild)
	}

	if err := handler.SyncCommands(s.Client, cmds, guildIDs); err != nil {
		return fmt.Errorf("error while syncing commands: %w", err)
	}
	return nil
}

func (s *Session) Open(ctx context.Context) error {
	if err := s.Client.OpenShardManager(ctx); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}
	return nil
}

func (s *Session) Close() error {
	s.Client.Close(context.Background())
	return nil
}
