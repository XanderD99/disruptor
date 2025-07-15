package disruptor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/snowflake/v2"
)

// Session represents the Discord bot
type Session struct {
	bot.Client
	handler.Router
}

// New creates a new Discord bot with sharding support
func New(token string, opts ...bot.ConfigOpt) (*Session, error) {
	router := handler.New()
	router.Use(
		middleware.GoErrDefer(
			func(e *handler.InteractionEvent, err error) {
				e.Client().Logger().Error("Error handling interaction", slog.Any("error", err))
				_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
					Embeds: &[]discord.Embed{
						{
							Title:       "Error",
							Description: fmt.Sprintf("An error occurred while processing your request: %s", err.Error()),
							Color:       0xFF0000, // Red color for error
						},
					},
				})
				if err != nil {
					e.Client().Logger().Error("Failed to update interaction response", slog.Any("error", err))
				}
			},
			discord.InteractionTypeApplicationCommand,
			false,
			false,
		),
		middleware.Logger,
	)
	options := []bot.ConfigOpt{}
	options = append(options, bot.WithEventListeners(router))
	options = append(options, opts...)

	c, err := disgo.New(token, options...)
	if err != nil {
		return nil, err
	}

	s := &Session{
		Client: c,
		Router: router,
	}

	return s, nil
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
