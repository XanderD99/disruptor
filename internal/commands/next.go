package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/internal/scheduler/handlers"
	"github.com/XanderD99/disruptor/pkg/logging"
)

type next struct {
	manager *scheduler.Manager
	db      *bun.DB
}

func Next(db *bun.DB, manager *scheduler.Manager) disruptor.Command {
	return next{manager: manager, db: db}
}

// Load implements disruptor.Command.
func (n next) Load(r handler.Router) {
	r.SlashCommand("/next", n.handle)
}

// Options implements disruptor.Command.
func (n next) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "next",
		Description: "Get next interval time.",
	}
}

func (n next) handle(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	// Get logger from context (added by the middleware)
	logger := logging.FromContext(e.Ctx)

	guildID := e.GuildID()
	if guildID == nil {
		return fmt.Errorf("guild ID is required for this command")
	}

	guild := models.Guild{ID: *guildID}
	if err := n.db.NewSelect().Model(&guild).WherePK().Scan(e.Ctx, &guild); err != nil {
		return fmt.Errorf("failed to find guild: %w", err)
	}

	logger.DebugContext(e.Ctx, "looking up scheduler for guild", "interval", guild.Interval)

	group, ok := n.manager.GetScheduler(handlers.HandlerTypeRandomVoiceJoin, guild.Interval)
	if !ok {
		logger.WarnContext(e.Ctx, "no scheduler found for interval", "interval", guild.Interval)
		return fmt.Errorf("no scheduler found")
	}

	interval := group.NextInterval()
	logger.DebugContext(e.Ctx, "retrieved next interval time", "next_time", interval)

	embed := discord.Embed{
		Description: fmt.Sprintf("Next interval time: <t:%d:F> <t:%d:R>", interval.Unix(), interval.Unix()),
		Color:       0x5c5fea, // PrimaryColor
	}

	_, err := e.UpdateInteractionResponse(discord.NewMessageUpdateBuilder().AddEmbeds(embed).Build())
	if err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*next)(nil)
