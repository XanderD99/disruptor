package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/discord-disruptor/internal/disruptor"
	"github.com/XanderD99/discord-disruptor/internal/scheduler"
)

type next struct {
	manager scheduler.Manager
}

func Next(manager scheduler.Manager) disruptor.Command {
	return next{
		manager: manager,
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
	guild := e.GuildID()
	if guild == nil {
		return fmt.Errorf("guild ID is required for this command")
	}

	group, err := p.manager.GetSchedulerForGuild(guild.String())
	if err != nil {
		return fmt.Errorf("failed to get interval group: %w", err)
	}

	interval := group.GetNextIntervalTime()

	embed := discord.Embed{
		Description: fmt.Sprintf("Next interval time: <t:%d:F> <t:%d:R>", interval.Unix(), interval.Unix()),
		Color:       0x5c5fea, // PrimaryColor
	}

	_, err = e.UpdateInteractionResponse(discord.NewMessageUpdateBuilder().AddEmbeds(embed).Build())
	if err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}
	return nil
}

var _ disruptor.Command = (*next)(nil)
