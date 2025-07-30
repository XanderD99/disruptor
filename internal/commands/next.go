package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/db"
)

type next struct {
	manager scheduler.Manager
	db      db.Database
}

func Next(manager scheduler.Manager, db db.Database) disruptor.Command {
	return next{
		manager: manager,
		db:      db,
	}
}

// Load implements disruptor.Command.
func (p next) Load(r handler.Router) {
	r.SlashCommand("/next", p.handle)
}

// Options implements disruptor.Command.
func (p next) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "next",
		Description: "Get next interval time.",
	}
}

func (p next) handle(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	guildID := e.GuildID()
	if guildID == nil {
		return fmt.Errorf("guild ID is required for this command")
	}

	var guild models.Guild
	if err := p.db.FindByID(e.Ctx, *guildID, &guild); err != nil {
		return fmt.Errorf("failed to find guild: %w", err)
	}

	group, ok := p.manager.GetScheduler(guild.Interval)
	if !ok {
		return fmt.Errorf("no scheduler found")
	}

	interval := group.GetNextIntervalTime()

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
