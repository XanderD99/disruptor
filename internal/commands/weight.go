package commands

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/omit"
	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/util"
)

type weight struct {
	db *bun.DB
}

func Weight(db *bun.DB) disruptor.Command {
	return weight{db: db}
}

// Load implements disruptor.Command.
func (w weight) Load(r handler.Router) {
	r.SlashCommand("/weight", w.handle)
}

var (
	minWeight = 0
	maxWeight = 100
)

// Options implements disruptor.Command.
func (w weight) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:                     "weight",
		Description:              "Changes the chance of a channel being selected for random voice joins.",
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionManageGuild),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:         "channel",
				Description:  "The channel to set the weight for",
				Required:     true,
				ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildVoice},
			},
			discord.ApplicationCommandOptionInt{
				Name:        "weight",
				Description: "The weight to set for the channel (between 0 and 100). ",
				MinValue:    &minWeight,
				MaxValue:    &maxWeight,
			},
		},
	}
}

func (w weight) handle(d discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	channel := d.Channel("channel")

	guildID := event.GuildID()
	if guildID == nil {
		return fmt.Errorf("this command can only be used in a guild")
	}

	model := models.DefaultChannel(channel.ID, *guildID)
	if err := w.db.NewSelect().Model(model).WherePK().Scan(event.Ctx, model); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to get channel %s from database: %w", channel.ID, err)
	}

	weight, ok := d.OptInt("weight")
	if !ok {
		embed := discord.NewEmbedBuilder()
		embed.SetColor(util.RGBToInteger(255, 215, 0))

		embed.SetDescription(fmt.Sprintf("Current weight for <#%d>: %.0f", channel.ID, model.Weight*100.0))

		msg := discord.NewMessageUpdateBuilder().SetEmbeds((embed).Build()).Build()

		if _, err := event.UpdateInteractionResponse(msg); err != nil {
			return fmt.Errorf("failed to update interaction response: %w", err)
		}

		return nil
	}

	model.Weight = float64(weight) / 100.0 // Scale to 0.0 - 100.0

	if _, err := w.db.NewInsert().Model(model).On("CONFLICT (id) DO UPDATE").Exec(event.Ctx); err != nil {
		return fmt.Errorf("failed to update channel %s in database: %w", channel.ID, err)
	}

	embed := discord.NewEmbedBuilder()
	embed.SetColor(util.RGBToInteger(0, 255, 0))
	embed.SetDescription(fmt.Sprintf("Set weight for <#%d> to %.0f", channel.ID, float64(weight)))

	msg := discord.NewMessageUpdateBuilder().SetEmbeds((embed).Build()).Build()

	if _, err := event.UpdateInteractionResponse(msg); err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*play)(nil)
