package gomigrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/korfairo/migratory/internal/gomigrator/dialect"
)

type Store interface {
	MigrationsTableExists(ctx context.Context, db DB) (bool, error)
	CreateMigrationsTable(ctx context.Context, db DB) error
	InsertMigration(ctx context.Context, db DB, migrationName string, id int64) error
	DeleteMigration(ctx context.Context, db DB, id int64) error
	SelectLastID(ctx context.Context, db DB) (int64, error)
	ListMigrations(ctx context.Context, db DB) ([]MigrationResult, error)
}

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type MigrationResult struct {
	ID        int64
	Name      string
	AppliedAt time.Time
}

const (
	DialectPostgres = "postgres"
)

var ErrUnsupportedDialect = errors.New("unsupported dialect")

type migrationStore struct {
	schemaName string
	tableName  string

	queryManager dialect.QueryManager
}

func newStore(dbDialect, schemaName, tableName string) (Store, error) {
	var queryManager dialect.QueryManager

	switch dbDialect {
	case DialectPostgres:
		queryManager = &dialect.Postgres{}
	default:
		return nil, ErrUnsupportedDialect
	}

	return migrationStore{
		queryManager: queryManager,
		schemaName:   schemaName,
		tableName:    tableName,
	}, nil
}

func (s migrationStore) MigrationsTableExists(ctx context.Context, db DB) (bool, error) {
	q := s.queryManager.MigrationsTableExists(s.schemaName, s.tableName)
	row := db.QueryRowContext(ctx, q)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to scan row: %w", err)
	}

	return exists, nil
}

func (s migrationStore) CreateMigrationsTable(ctx context.Context, db DB) error {
	q := s.queryManager.CreateMigrationsTable(s.schemaName, s.tableName)
	_, err := db.ExecContext(ctx, q)
	return err
}

func (s migrationStore) InsertMigration(ctx context.Context, db DB, migrationName string, id int64) error {
	q := s.queryManager.InsertMigration(s.schemaName, s.tableName)
	_, err := db.ExecContext(ctx, q, id, migrationName)
	return err
}

func (s migrationStore) DeleteMigration(ctx context.Context, db DB, id int64) error {
	q := s.queryManager.DeleteMigration(s.schemaName, s.tableName)
	_, err := db.ExecContext(ctx, q, id)
	return err
}

var ErrNoRows = errors.New("no rows in migrations table")

func (s migrationStore) SelectLastID(ctx context.Context, db DB) (int64, error) {
	q := s.queryManager.SelectLastMigrationID(s.schemaName, s.tableName)
	row := db.QueryRowContext(ctx, q)

	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, ErrNoRows
		}
		return -1, fmt.Errorf("failed to scan migration id: %w", err)
	}
	if err := row.Err(); err != nil {
		return -1, fmt.Errorf("an error occurred during row scanning: %w", err)
	}

	return id, nil
}

func (s migrationStore) ListMigrations(ctx context.Context, db DB) ([]MigrationResult, error) {
	q := s.queryManager.ListMigrations(s.schemaName, s.tableName)
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query ListMigrations: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var migrations []MigrationResult
	for rows.Next() {
		var id int64
		var name string
		var appliedAt time.Time
		if err = rows.Scan(&id, &name, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration result: %w", err)
		}
		migrations = append(migrations, MigrationResult{
			ID:        id,
			Name:      name,
			AppliedAt: appliedAt,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred during iteration through sql rows: %w", err)
	}

	return migrations, nil
}
