package commands

import (
	"fmt"

	"github.com/XanderD99/discord-disruptor/internal/disruptor"
	"github.com/XanderD99/discord-disruptor/internal/models"
	"github.com/XanderD99/discord-disruptor/pkg/database"
	"github.com/XanderD99/discord-disruptor/pkg/util"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/sync/errgroup"
)

type soundboard struct {
	db database.Database
}

func SoundBoard(db database.Database) disruptor.Command {
	return soundboard{
		db: db,
	}
}

// Load implements disruptor.Command.
func (p soundboard) Load(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Route("/soundboard", func(r handler.Router) {
			r.Command("/toggle", p.toggle)
			r.Autocomplete("/toggle", p.Autocomplete)
			r.Command("/list", p.list)
		})
	},
	)
}

// Options implements disruptor.Command.
func (p soundboard) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "soundboard",
		Description: "Manage soundboard settings",
		Options: []discord.ApplicationCommandOption{

			discord.ApplicationCommandOptionSubCommand{
				Name:        "toggle",
				Description: "Toggle a soundboard sound",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "sound",
						Description:  "The soundboard sound to toggle",
						Required:     true,
						Autocomplete: true,
					},
					discord.ApplicationCommandOptionBool{
						Name:        "enabled",
						Description: "Whether the sound should be enabled or disabled",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "list",
				Description: "List all soundboard sounds and their status",
			},
		},
	}
}

func (p soundboard) Autocomplete(e *handler.AutocompleteEvent) error {
	guild := e.GuildID()
	if guild == nil {
		return fmt.Errorf("guild ID is required for this command")
	}

	client := e.Client()

	choices := make([]discord.AutocompleteChoice, 0)
	client.Caches().GuildSoundboardSoundsForEach(*guild, func(sound discord.SoundboardSound) {
		choices = append(choices, discord.AutocompleteChoiceString{
			Name:  sound.Name,
			Value: sound.SoundID.String(),
		})
	})

	if err := e.AutocompleteResult(choices); err != nil {
		return fmt.Errorf("failed to send autocomplete result: %w", err)
	}
	return nil
}

func (p soundboard) toggle(e *handler.CommandEvent) error {
	client := e.Client()
	guild := e.GuildID()
	if guild == nil {
		return fmt.Errorf("guild ID is required for this command")
	}

	soundSnowflake, err := snowflake.Parse(e.SlashCommandInteractionData().String("sound"))
	if err != nil {
		return fmt.Errorf("invalid sound ID: %w", err)
	}

	enabled := e.SlashCommandInteractionData().Bool("enabled")

	data, err := p.db.FindByID(e.Ctx, soundSnowflake.String(), models.Sound{})
	if err != nil {

		s, ok := client.Caches().GuildSoundboardSound(*guild, soundSnowflake)
		if !ok {
			return fmt.Errorf("sound (id: %s) not found in guild (id: %s)", soundSnowflake, guild.String())
		}

		data = &models.Sound{ID: soundSnowflake, Name: s.Name, URL: s.URL(), GuildID: *guild}
	}

	sound, ok := data.(*models.Sound)
	if !ok {
		return fmt.Errorf("expected models.Sound, got %T", data)
	}

	if sound.Enabled == enabled {
		return fmt.Errorf("sound %s is already %s", sound.Name, map[bool]string{
			true:  "enabled",
			false: "disabled",
		}[enabled])
	}

	if sound.GuildID != *guild {
		return fmt.Errorf("sound %s is not available in this guild", sound.Name)
	}

	sound.Enabled = enabled
	if err := p.db.Upsert(e.Ctx, sound); err != nil {
		return fmt.Errorf("failed to update sound %s: %w", sound.Name, err)
	}

	_, err = e.UpdateInteractionResponse(discord.NewMessageUpdateBuilder().SetContent("Sound enabled").Build())
	if err != nil {
		return fmt.Errorf("failed to update interaction response: %w", err)
	}

	return nil
}

func (p soundboard) list(e *handler.CommandEvent) error {
	client := e.Client()
	guild := e.GuildID()
	if guild == nil {
		return fmt.Errorf("guild ID is required for this command")
	}

	filter := map[string]any{
		"$or": []any{
			map[string]any{"guild_id": guild.String()},
			map[string]any{"guild_id": nil},
		},
	}

	data, err := p.db.FindAll(e.Ctx, models.Sound{}, database.WithFilters(filter), database.WithSort([]database.Sort{{Field: "global", Direction: database.Asc}}))
	if err != nil {
		return fmt.Errorf("failed to retrieve soundboard sounds: %w", err)
	}

	sounds, ok := data.([]models.Sound)
	if !ok {
		return fmt.Errorf("expected []models.Sound, got %T", data)
	}

	if len(sounds) == 0 {
		return fmt.Errorf("no soundboard sounds found for guild %s", guild.String())
	}

	// split between guild and global sounds
	var guildSounds, globalSounds []models.Sound
	for _, sound := range sounds {
		if sound.Global {
			globalSounds = append(globalSounds, sound)
		} else {
			guildSounds = append(guildSounds, sound)
		}
	}

	buildEmbedField := func(model models.Sound) discord.EmbedField {
		inline := false

		enbledEmojis := map[bool]string{
			true:  "✅",
			false: "❌",
		}

		return discord.EmbedField{
			Name:   fmt.Sprintf("%s (%s)", model.Name, model.ID),
			Value:  fmt.Sprintf("Enabled: %s", enbledEmojis[model.Enabled]),
			Inline: &inline,
		}
	}

	const maxFieldsPerEmbed = 25  // Discord limit is 25 fields per embed
	const maxEmbedPerMessage = 10 // Discord limit is 10 embeds per message

	buildAndSendEmbeds := func(title string, sounds []models.Sound) error {
		fields := make([]discord.EmbedField, 0, len(sounds))
		for _, sound := range sounds {
			fields = append(fields, buildEmbedField(sound))
		}

		fieldChunks := util.Chunk(fields, maxFieldsPerEmbed)

		embeds := make([]discord.Embed, 0, len(fieldChunks))
		for i, chunk := range fieldChunks {
			embed := discord.Embed{
				Color:  0x5c5fea, // PrimaryColor
				Fields: chunk,
			}

			if i == 0 {
				embed.Title = title
			}

			embeds = append(embeds, embed)
		}

		embedChunks := util.Chunk(embeds, maxEmbedPerMessage)
		var eg errgroup.Group

		for _, chunk := range embedChunks {
			eg.Go(func() error {
				_, err := client.Rest().CreateMessage(e.Channel().ID(), discord.NewMessageCreateBuilder().SetEmbeds(chunk...).Build())
				return err
			})
		}

		return eg.Wait()
	}

	if len(globalSounds) > 0 {
		if err := buildAndSendEmbeds("Global Sounds", globalSounds); err != nil {
			return fmt.Errorf("failed to send global sounds: %w", err)
		}
	}

	if len(guildSounds) > 0 {
		if err := buildAndSendEmbeds(fmt.Sprintf("Guild Sounds (%s)", guild.String()), guildSounds); err != nil {
			return fmt.Errorf("failed to send guild sounds: %w", err)
		}
	}

	if err := e.DeleteInteractionResponse(); err != nil {
		return fmt.Errorf("failed to delete interaction response: %w", err)
	}

	return nil
}

var _ disruptor.Command = (*soundboard)(nil)
