package gomigrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

var (
	ErrUnknownDBVersion  = errors.New("migration table was not created before, database version is unknown")
	ErrDirtyMigrations   = errors.New("dirty migration(s) found (unapplied one with ID less than database version)")
	ErrNothingToRollback = errors.New("nothing to rollback, no rows in migrations table")
)

type Migrator struct {
	store Store
}

func New(dialect, schemaName, tableName string) (*Migrator, error) {
	store, err := newStore(dialect, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	return &Migrator{store}, nil
}

func (m Migrator) Up(ctx context.Context, migrations Migrations, db *sql.DB, force bool) (n int, err error) {
	exists, err := m.ensureMigrationTableExists(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure that migration table exists: %w", err)
	}

	if !exists {
		return m.upAll(ctx, migrations, db)
	}

	dbMigrations, err := m.store.ListMigrations(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("failed to get migrations results from database: %w", err)
	}

	missingMigrations, dirty := findMissingMigrations(migrations, dbMigrations)
	if !force && dirty {
		return 0, ErrDirtyMigrations
	}

	return m.upAll(ctx, missingMigrations, db)
}

func (m Migrator) Down(ctx context.Context, migrations Migrations, db *sql.DB, redo bool) error {
	return m.downLast(ctx, migrations, db, redo)
}

func (m Migrator) GetStatus(ctx context.Context, migrations Migrations, db *sql.DB) ([]MigrationResult, error) {
	_, err := m.ensureMigrationTableExists(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure that migration table exists: %w", err)
	}

	dbMigrations, err := m.store.ListMigrations(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations results from database: %w", err)
	}

	missingMigrations, _ := findMissingMigrations(migrations, dbMigrations)

	results := append(make([]MigrationResult, 0, len(dbMigrations)+len(missingMigrations)), dbMigrations...)
	for _, missing := range missingMigrations {
		results = append(results, MigrationResult{
			ID:   missing.ID(),
			Name: missing.Name(),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return results, nil
}

func (m Migrator) GetDBVersion(ctx context.Context, db *sql.DB) (int64, error) {
	exists, err := m.ensureMigrationTableExists(ctx, db)
	if err != nil {
		return -1, fmt.Errorf("failed to ensure that migration table exists: %w", err)
	}

	if !exists {
		return -1, ErrUnknownDBVersion
	}

	lastVersion, err := m.store.SelectLastID(ctx, db)
	if err != nil {
		return -1, fmt.Errorf("failed to get current database version: %w", err)
	}

	return lastVersion, nil
}

func (m Migrator) upOne(ctx context.Context, migration Migration, db *sql.DB) error {
	noTx, err := migration.ChooseExecutor()
	if err != nil {
		return fmt.Errorf("failed to migration.ChooseExecutor(): %w", err)
	}

	if noTx {
		return m.upNoTx(ctx, migration, db)
	}

	return m.up(ctx, migration, db)
}

func (m Migrator) up(ctx context.Context, migration Migration, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = migration.Up(ctx, tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("failed to up migration and rollback transaction: %w; %w", err, txErr)
		}
		return fmt.Errorf("failed to up migration: %w", err)
	}

	if err = m.store.InsertMigration(ctx, tx, migration.Name(), migration.ID()); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("failed to insert migration in table and rollback transaction: %w; %w", err, txErr)
		}
		return fmt.Errorf("failed to insert migration in table: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m Migrator) upNoTx(ctx context.Context, migration Migration, db *sql.DB) error {
	if err := migration.UpNoTx(ctx, db); err != nil {
		return fmt.Errorf("failed to up migration: %w", err)
	}

	if err := m.store.InsertMigration(ctx, db, migration.Name(), migration.ID()); err != nil {
		return fmt.Errorf("failed to insert migration in table: %w", err)
	}

	return nil
}

func (m Migrator) downOne(ctx context.Context, migration *Migration, db *sql.DB) error {
	noTx, err := migration.ChooseExecutor()
	if err != nil {
		return fmt.Errorf("failed to migration.ChooseExecutor(): %w", err)
	}

	if noTx {
		return m.downNoTx(ctx, migration, db)
	}

	return m.down(ctx, migration, db)
}

func (m Migrator) down(ctx context.Context, migration *Migration, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = migration.Down(ctx, tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("failed to down migration and rollback transaction: %w; %w", err, txErr)
		}
		return fmt.Errorf("failed to up migration: %w", err)
	}

	if err = m.store.DeleteMigration(ctx, tx, migration.ID()); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("failed to delete migration in table and rollback transaction: %w; %w", err, txErr)
		}
		return fmt.Errorf("failed to insert migration in table: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m Migrator) downNoTx(ctx context.Context, migration *Migration, db *sql.DB) error {
	if err := migration.DownNoTx(ctx, db); err != nil {
		return fmt.Errorf("failed to down migration: %w", err)
	}

	if err := m.store.DeleteMigration(ctx, db, migration.ID()); err != nil {
		return fmt.Errorf("failed to delete migration from in table: %w", err)
	}

	return nil
}

func (m Migrator) upAll(ctx context.Context, migrations Migrations, db *sql.DB) (int, error) {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID() < migrations[j].ID()
	})

	var appliedCount int
	for _, migration := range migrations {
		if err := m.upOne(ctx, migration, db); err != nil {
			return appliedCount, fmt.Errorf("failed to up migration with ID %d: %w", migration.ID(), err)
		}
		appliedCount++
	}

	return appliedCount, nil
}

func (m Migrator) downLast(ctx context.Context, migrations Migrations, db *sql.DB, redo bool) error {
	exists, err := m.ensureMigrationTableExists(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to ensure that migration table exists: %w", err)
	}

	if !exists {
		return ErrUnknownDBVersion
	}

	last, err := m.getLastMigration(ctx, migrations, db)
	if err != nil {
		if errors.Is(err, ErrNoRows) {
			return ErrNothingToRollback
		}
		return fmt.Errorf("failed to find last migration: %w", err)
	}

	if err = m.downOne(ctx, last, db); err != nil {
		return fmt.Errorf("failed to rollback last migration with ID %d: %w", last.ID(), err)
	}

	if redo {
		if err = m.upOne(ctx, *last, db); err != nil {
			return fmt.Errorf("failed to apply last migration with ID %d: %w", last.ID(), err)
		}
	}

	return nil
}

func (m Migrator) ensureMigrationTableExists(ctx context.Context, db *sql.DB) (exists bool, err error) {
	exists, err = m.store.MigrationsTableExists(ctx, db)
	if err != nil {
		return false, fmt.Errorf("failed to check if migration table exists: %w", err)
	}

	if exists {
		return true, nil
	}

	if err = m.store.CreateMigrationsTable(ctx, db); err != nil {
		return false, fmt.Errorf("failed to create migration table: %w", err)
	}

	return false, nil
}

func (m Migrator) getLastMigration(ctx context.Context, ms Migrations, db *sql.DB) (*Migration, error) {
	lastID, err := m.store.SelectLastID(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get current database version: %w", err)
	}

	for _, migration := range ms {
		if migration.ID() == lastID {
			return &migration, nil
		}
	}

	return nil, fmt.Errorf("database version is %d, but migration with this ID not found", lastID)
}

func findMissingMigrations(migrations Migrations, results []MigrationResult) (missing Migrations, dirty bool) {
	appliedIds := make(map[int64]struct{}, len(results))
	var maxID int64

	for _, r := range results {
		appliedIds[r.ID] = struct{}{}
		if r.ID > maxID {
			maxID = r.ID
		}
	}

	for _, migration := range migrations {
		if _, exists := appliedIds[migration.ID()]; exists {
			continue
		}
		if migration.ID() < maxID {
			dirty = true
		}
		missing = append(missing, migration)
	}

	return missing, dirty
}
