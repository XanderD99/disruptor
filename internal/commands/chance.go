package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/pkg/db"
	"github.com/XanderD99/disruptor/pkg/util"
)

type chance struct {
	db db.Database
}

func Chance(db db.Database) disruptor.Command {
	return chance{
		db: db,
	}
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
		DefaultMemberPermissions: json.NewNullablePtr(discord.PermissionManageGuild),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "percentage",
				Description: "Percentage chance of the event occurring (0-100)",
			},
		},
	}
}

func (c chance) handle(d discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	ctx, cancel := context.WithCancel(event.Ctx)
	defer cancel()

	guildID := event.GuildID()
	if guildID == nil {
		return fmt.Errorf("this command can only be used in a guild")
	}
	var guild models.Guild
	err := c.db.FindByID(event.Ctx, *guildID, &guild)
	if err != nil {
		event.Client().Logger().Error("Failed to find guild", slog.Any("error", err))
		guild = models.NewGuild(*guildID)
	}

	percentage, ok := d.OptInt("percentage")
	if !ok {
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

	guild.Chance = models.Chance(percentage)

	if err := c.db.Upsert(ctx, guild); err != nil {
		return fmt.Errorf("failed to update guild chance: %w", err)
	}

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
