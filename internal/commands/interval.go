package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/models"
	"github.com/XanderD99/disruptor/internal/scheduler"
	"github.com/XanderD99/disruptor/pkg/database"
	"github.com/XanderD99/disruptor/pkg/util"
)

type interval struct {
	manager scheduler.Manager
	store   database.Database
}

func Interval(store database.Database, manager scheduler.Manager) disruptor.Command {
	return interval{
		manager: manager,
		store:   store,
	}
}

// Load implements disruptor.Command.
func (i interval) Load(r handler.Router) {
	r.SlashCommand("/interval", i.handle)
}

// Options implements disruptor.Command.
func (i interval) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:                     "interval",
		Description:              "Set the interval for the audio scheduler",
		DefaultMemberPermissions: json.NewNullablePtr(discord.PermissionManageGuild),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "duration",
				Description: "Duration of the interval, Valid time units are \"s\", \"m\", \"h\". example: 300m, 1.5h or 2h45m",
			},
		},
	}
}

func (i interval) handle(d discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	guildID := event.GuildID()
	if guildID == nil {
		return fmt.Errorf("this command can only be used in a guild")
	}

	data, err := i.store.FindByID(ctx, guildID.String(), &models.Guild{})
	if err != nil {
		event.Client().Logger().Error("Failed to find guild", slog.Any("error", err))
		data = models.NewGuild(*guildID)
	}

	if data == nil {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	guild, ok := data.(*models.Guild)
	if !ok {
		return fmt.Errorf("failed to cast guild data: %T", data)
	}

	intervalString, ok := d.OptString("duration")
	if !ok {

		embed := discord.NewEmbedBuilder()
		embed.SetColor(util.RGBToInteger(255, 215, 0))

		embed.SetDescription(fmt.Sprintf("Current interval: %s", guild.Interval))

		msg := discord.NewMessageUpdateBuilder().SetEmbeds((embed).Build()).Build()

		if _, err := event.UpdateInteractionResponse(msg); err != nil {
			return fmt.Errorf("failed to update interaction response: %w", err)
		}

		return nil
	}

	guild.Interval, err = time.ParseDuration(intervalString)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	if guild.Interval < (time.Minute * 10) {
		return fmt.Errorf("invalid duration: %s, must be greater or equal than 10m", intervalString)
	}

	if guild.Interval > (time.Hour * 24) {
		return fmt.Errorf("invalid duration: %s, must be less than 24h", intervalString)
	}

	if err := i.store.Upsert(ctx, guild); err != nil {
		return fmt.Errorf("failed to update guild interval: %w", err)
	}

	if err := i.manager.AddGuild(guildID.String(), guild.Interval); err != nil {
		return fmt.Errorf("failed to add guild to manager: %w", err)
	}

	embed := discord.NewEmbedBuilder()
	embed.SetColor(util.RGBToInteger(255, 215, 0))
	embed.SetDescription(fmt.Sprintf("Interval set to: %s", guild.Interval))
	msg := discord.NewMessageUpdateBuilder().SetEmbeds(embed.Build()).Build()
	if _, err := event.UpdateInteractionResponse(msg); err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*play)(nil)
