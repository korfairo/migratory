package command

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/korfairo/migratory/internal/gomigrator"
	"github.com/korfairo/migratory/internal/sqlmigration"
)

func rollback(dir, schema, table string, redo bool) error {
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}

	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("failed to close database connection")
		}
	}()

	migrations, err := sqlmigration.SeekMigrations(dir, nil)
	if err != nil {
		return fmt.Errorf("could not find migrations in directory %s: %w", dir, err)
	}

	ctx := context.Background()
	migrator, err := gomigrator.New(ctx, db, "postgres", schema, table)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err = migrator.Down(ctx, migrations, db, redo); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
