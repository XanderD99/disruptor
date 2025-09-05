package models

import (
	"github.com/disgoorg/snowflake/v2"
)

func DefaultChannel(id, guildID snowflake.ID) *Channel {
	return &Channel{ID: id, GuildID: guildID, Weight: .5}
}

type Channel struct {
	ID snowflake.ID `bun:"id,pk" validate:"required"` // snowflake ID of the guild

	Guild   Guild        `bun:"rel:belongs-to,join:guild_id=id"` // the guild this channel belongs to
	GuildID snowflake.ID `bun:"guild_id" validate:"required"`    // snowflake ID of the guild

	Weight float64 `bun:"weight,notnull,default:.5"` // weight for selection, default .5
}
