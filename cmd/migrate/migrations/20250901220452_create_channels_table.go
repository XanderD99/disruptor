package migrations

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/XanderD99/disruptor/internal/models"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.NewCreateTable().Model((*models.Channel)(nil)).IfNotExists().Exec(ctx)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.NewDropTable().Model((*models.Channel)(nil)).IfExists().Exec(ctx)
		return err
	})
}
