package models

import (
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

const (
	defaultInterval = time.Hour
	defaultChance   = 40
)

func NewGuild(snowflake snowflake.ID) Guild {
	return Guild{
		ID:       snowflake,
		Interval: defaultInterval,
		Chance:   defaultChance,
	}
}

type Guild struct {
	ID       snowflake.ID  `bun:"id,pk" validate:"required"`               // snowflake ID of the guild
	Chance   Chance        `bun:"chance" validate:"required,gt=0,lte=100"` // chance of a sound being played
	Interval time.Duration `bun:"interval" validate:"required"`            // interval between sounds

	Channels []Channel `bun:"rel:has-many,join:id=guild_id"` // channels in the guild
}

type Chance int

func (c Chance) String() string {
	return fmt.Sprintf("%d%%", c)
}
