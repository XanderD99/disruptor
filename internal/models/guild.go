package models

import (
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func NewGuild(id snowflake.ID) *Guild {
	return &Guild{
		ID: id,
		Settings: GuildSettings{
			Interval: time.Hour,
			Chance:   40,
		},
	}
}

type Guild struct {
	ID snowflake.ID `json:"id" bson:"id" validate:"required"` // snowflake ID of the guild

	Settings GuildSettings `json:"settings" bson:"settings"`
}

type Chance int

func (c Chance) String() string {
	return fmt.Sprintf("%d%%", c)
}

type GuildSettings struct {
	Chance   Chance        `json:"chance" bson:"chance" validate:"required,gt=0,lt=1"` // chance of a sound being played
	Interval time.Duration `json:"interval" bson:"interval" validate:"required"`       // interval between sounds
}
