package commands

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"

	"github.com/XanderD99/discord-disruptor/internal/disruptor"
	"github.com/XanderD99/discord-disruptor/internal/store"
	"github.com/XanderD99/discord-disruptor/pkg/util"
)

type chance struct {
	store store.Store
}

func Chance(store store.Store) disruptor.Command {
	return chance{
		store: store,
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	guild, err := c.store.Guilds().FindByID(ctx, event.GuildID().String())
	if err != nil {
		return fmt.Errorf("failed to find guild: %w", err)
	}

	percentage, ok := d.OptInt("percentage")
	if !ok {
		embed := discord.NewEmbedBuilder()
		embed.SetColor(util.RGBToInteger(255, 215, 0))

		embed.SetDescription(fmt.Sprintf("Current chance percentage: %d%%", int(guild.Settings.Chance*100)))

		msg := discord.NewMessageUpdateBuilder().SetEmbeds((embed).Build()).Build()

		if _, err := event.UpdateInteractionResponse(msg); err != nil {
			return fmt.Errorf("failed to update interaction response: %w", err)
		}

		return nil
	}

	if (percentage / 100) == int(guild.Settings.Chance) {
		embed := discord.NewEmbedBuilder()
		embed.SetColor(util.RGBToInteger(255, 215, 0))
		embed.SetDescription(fmt.Sprintf("Chance is already set to: %d%%", percentage))
		msg := discord.NewMessageUpdateBuilder().SetEmbeds(embed.Build()).Build()
		if _, err := event.UpdateInteractionResponse(msg); err != nil {
			return fmt.Errorf("failed to update interaction response: %w", err)
		}
		return nil
	}

	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	guild.Settings.Chance = float64(percentage / 100)
	if err := c.store.Guilds().Update(ctx, guild.ID, guild); err != nil {
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

var _ disruptor.Command = (*play)(nil)
