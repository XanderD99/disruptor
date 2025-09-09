package disruptor

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

// Disruptor represents the Discord bot
type Disruptor struct {
	*bot.Client
}

type opts struct {
	commands    []Command
	middlewares []handler.Middleware
}

type optFunc func(*opts)

func WithMiddlewares(middlewares ...handler.Middleware) optFunc {
	return func(o *opts) {
		o.middlewares = append(o.middlewares, middlewares...)
	}
}

func WithCommands(commands ...Command) optFunc {
	return func(o *opts) {
		o.commands = commands
	}
}

// New creates a new Discord bot with sharding support
func New(cfg Config, optFuncs ...optFunc) (*Disruptor, error) {
	opts := new(opts)
	for _, o := range optFuncs {
		o(opts)
	}

	router := handler.New()
	router.Use(opts.middlewares...)

	options := []bot.ConfigOpt{}
	options = append(options, bot.WithEventListeners(router))
	options = append(options, cfg.ToSessionOpts()...)

	c, err := disgo.New(cfg.Token, options...)
	if err != nil {
		return nil, err
	}

	cmds := make([]discord.ApplicationCommandCreate, 0, len(opts.commands))
	for _, cmd := range opts.commands {
		if cmd == nil {
			continue // Skip nil commands
		}
		cmd.Load(router)
		cmds = append(cmds, cmd.Options())
	}

	guild := snowflake.GetEnv("CONFIG_GUILDID")
	guildIDs := []snowflake.ID{}
	if guild != 0 {
		guildIDs = append(guildIDs, guild)
	}

	if err := handler.SyncCommands(c, cmds, guildIDs); err != nil {
		return nil, fmt.Errorf("error while syncing commands: %w", err)
	}

	return &Disruptor{Client: c}, nil
}

func (s *Disruptor) Open(ctx context.Context) error {
	if err := s.Client.OpenShardManager(ctx); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}
	return nil
}

func (s *Disruptor) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Client.Close(ctx)
	<-ctx.Done()

	return ctx.Err()
}
