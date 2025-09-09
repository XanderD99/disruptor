package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/omit"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/util"
	"github.com/XanderD99/disruptor/pkg/logging"
)

type chance struct {
	db *bun.DB
}

func Chance(db *bun.DB) disruptor.Command {
	return chance{db: db}
}

// Load implements disruptor.Command.
func (c chance) Load(r handler.Router) {
	r.SlashCommand("/chance", c.handle)
}

// Options implements disruptor.Command.
func (c chance) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:                     "chance",
		Description:              "Set the chance of an event occurring",
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionManageGuild),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "percentage",
				Description: "Percentage chance of the event occurring (0-100)",
			},
		},
	}
}

func (c chance) handle(d discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	// Get logger from context (added by the middleware)
	logger := logging.FromContext(event.Ctx)

	guildID := event.GuildID()
	if guildID == nil {
		return fmt.Errorf("this command can only be used in a guild")
	}

	guild := models.Guild{ID: *guildID}
	if err := c.db.NewSelect().Model(&guild).WherePK().Scan(event.Ctx, &guild); err != nil {
		guild = models.NewGuild(*guildID)
	}

	percentage, ok := d.OptInt("percentage")
	if !ok {
		logger.DebugContext(event.Ctx, "displaying current chance percentage", "current_chance", guild.Chance)

		embed := discord.NewEmbedBuilder()
		embed.SetColor(util.RGBToInteger(255, 215, 0))

		embed.SetDescription(fmt.Sprintf("Current chance percentage: %s", guild.Chance))

		msg := discord.NewMessageUpdateBuilder().SetEmbeds((embed).Build()).Build()

		if _, err := event.UpdateInteractionResponse(msg); err != nil {
			return fmt.Errorf("failed to update interaction response: %w", err)
		}

		return nil
	}

	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	oldChance := guild.Chance
	guild.Chance = models.Chance(percentage)

	logger.DebugContext(event.Ctx, "updating guild chance", "old_chance", oldChance, "new_chance", guild.Chance)

	if _, err := c.db.NewUpdate().Model(&guild).WherePK().Exec(event.Ctx); err != nil {
		return fmt.Errorf("failed to update guild chance: %w", err)
	}

	logger.DebugContext(event.Ctx, "successfully updated guild chance", "new_chance", percentage)

	embed := discord.NewEmbedBuilder()
	embed.SetColor(util.RGBToInteger(255, 215, 0))
	embed.SetDescription(fmt.Sprintf("Chance set to: %d%%", percentage))
	msg := discord.NewMessageUpdateBuilder().SetEmbeds(embed.Build()).Build()
	if _, err := event.UpdateInteractionResponse(msg); err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*chance)(nil)
