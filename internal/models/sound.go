package models

import (
	"github.com/disgoorg/snowflake/v2"
)

type Sound struct {
	ID      snowflake.ID `json:"id" bson:"id" validate:"required"`             // snowflake ID of the sound
	GuildID snowflake.ID `json:"guild_id" bson:"guild_id" validate:"required"` // snowflake ID of the guild, empty if global sound

	Name    string `json:"name" bson:"name" validate:"required"`
	URL     string `json:"url" bson:"url" validate:"required,url"`
	Enabled bool   `json:"enabled" bson:"enabled"`
	Global  bool   `json:"global" bson:"global"` // true if sound is global, false if guild-specific
}
