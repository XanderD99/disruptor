package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/migrate"

	"github.com/XanderD99/disruptor/cmd/migrate/migrations"

	"github.com/urfave/cli/v2"
)

func migrator(dsn string) (*migrate.Migrator, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Add slog hook for logging (without custom metrics)
	return migrate.NewMigrator(db, migrations.Migrations), nil
}

func main() { //nolint:gocyclo
	app := &cli.App{
		Name: "migrate",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "connection",
				Aliases: []string{"c"},
				Usage:   "Database connection string",
				Value:   "file::memory:?cache=shared",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					m, err := migrator(c.String("connection"))
					if err != nil {
						return err
					}

					return m.Init(c.Context)
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					m, err := migrator(c.String("connection"))
					if err != nil {
						return err
					}

					if err := m.Lock(c.Context); err != nil {
						return err
					}
					defer m.Unlock(c.Context) //nolint:errcheck

					group, err := m.Migrate(c.Context)
					if err != nil {
						fmt.Printf("migration failed, rolling back: %v\n", err)
						group, err := m.Rollback(c.Context)
						if err != nil {
							return fmt.Errorf("rollback failed: %v", err)
						}
						if group.IsZero() {
							fmt.Printf("there are no groups to roll back\n")
							return err
						}
						return err
					}
					if group.IsZero() {
						fmt.Printf("there are no new migrations to run (database is up to date)\n")
						return nil
					}
					fmt.Printf("migrated to %s\n", group)
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					m, err := migrator(c.String("connection"))
					if err != nil {
						return err
					}

					if err := m.Lock(c.Context); err != nil {
						return err
					}
					defer m.Unlock(c.Context) //nolint:errcheck

					group, err := m.Rollback(c.Context)
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Printf("there are no groups to roll back\n")
						return nil
					}
					fmt.Printf("rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					m, err := migrator(c.String("connection"))
					if err != nil {
						return err
					}

					ms, err := m.MigrationsWithStatus(c.Context)
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())
					return nil
				},
			},
			{
				Name:  "mark_applied",
				Usage: "mark migrations as applied without actually running them",
				Action: func(c *cli.Context) error {
					m, err := migrator(c.String("connection"))
					if err != nil {
						return err
					}

					group, err := m.Migrate(c.Context, migrate.WithNopMigration())
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Printf("there are no new migrations to mark as applied\n")
						return nil
					}
					fmt.Printf("marked as applied %s\n", group)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
