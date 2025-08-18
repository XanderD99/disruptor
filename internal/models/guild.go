package models

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/uptrace/bun"
)

func NewGuild(snowflake snowflake.ID) Guild {
	return Guild{
		Snowflake: snowflake,
		Interval:  time.Hour,
		Chance:    40,
	}
}

type Guild struct {
	Snowflake snowflake.ID  `bun:"snowflake,pk" validate:"required"`        // snowflake ID of the guild
	Chance    Chance        `bun:"chance" validate:"required,gt=0,lte=100"` // chance of a sound being played
	Interval  time.Duration `bun:"interval" validate:"required"`            // interval between sounds

	CreatedAt time.Time `bun:"create_at,nullzero,default:current_timestamp" validate:"required"` // when the guild was created
	UpdatedAt time.Time `bun:"update_at,nullzero,default:current_timestamp" validate:"required"` // when the guild was last updated
	DeletedAt time.Time `bun:"deleted_at,soft_delete,nullzero" validate:"required"`              // when the guild was deleted
}

var _ bun.BeforeAppendModelHook = (*Guild)(nil)

func (m *Guild) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now()
	}
	return nil
}

type Chance int

func (c Chance) String() string {
	return fmt.Sprintf("%d%%", c)
}
