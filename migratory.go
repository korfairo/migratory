package migratory

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/korfairo/migratory/internal/gomigrator"
	"github.com/korfairo/migratory/internal/sqlmigration"
)

type options struct {
	migrationType string
	directory     string
	dialect       string
	schema        string
	table         string

	forceUp bool
}

type OptionsFunc func(o *options)

func applyOptions(optionsFns []OptionsFunc) options {
	opts := defaultOptions()
	for _, apply := range optionsFns {
		apply(&opts)
	}
	return opts
}

func WithGoMigration() OptionsFunc {
	return func(o *options) { o.migrationType = MigrationTypeGo }
}

func WithSQLMigrationDir(d string) OptionsFunc {
	return func(o *options) { o.migrationType = MigrationTypeSQL; o.directory = d }
}

func WithSchema(n string) OptionsFunc {
	return func(o *options) { o.schema = n }
}

func WithTable(n string) OptionsFunc {
	return func(o *options) { o.table = n }
}

func WithForce() OptionsFunc {
	return func(o *options) { o.forceUp = true }
}

var defaultOpts = options{
	migrationType: MigrationTypeGo,
	dialect:       DialectPostgres,
	directory:     ".",
	schema:        "public",
	table:         "migrations",
	forceUp:       false,
}

func defaultOptions() options {
	return defaultOpts
}

func SetSchema(s string) { defaultOpts.schema = s }

func SetTable(s string) { defaultOpts.table = s }

func SetSQLDirectory(path string) {
	defaultOpts.migrationType = MigrationTypeSQL
	defaultOpts.directory = path
}

const (
	MigrationTypeGo  = "go"
	MigrationTypeSQL = "sql"
)

const (
	DialectPostgres = "postgres"
)

func Up(db *sql.DB, opts ...OptionsFunc) (n int, err error) {
	ctx := context.Background()
	return UpContext(ctx, db, opts...)
}

func UpContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) (n int, err error) {
	option := applyOptions(opts)
	migrator, err := gomigrator.New(ctx, db, option.dialect, option.schema, option.table)
	if err != nil {
		return 0, err
	}

	migrations, err := getMigrations(option.migrationType, option.directory)
	if err != nil {
		return 0, err
	}

	appliedCount, err := migrator.Up(ctx, migrations, db, option.forceUp)
	if err != nil {
		return appliedCount, err
	}

	return appliedCount, nil
}

func Down(db *sql.DB, opts ...OptionsFunc) error {
	ctx := context.Background()
	return DownContext(ctx, db, opts...)
}

func DownContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) error {
	return rollback(ctx, db, false, opts)
}

func Redo(db *sql.DB, opts ...OptionsFunc) error {
	ctx := context.Background()
	return RedoContext(ctx, db, opts...)
}

func RedoContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) error {
	return rollback(ctx, db, true, opts)
}

type MigrationResult struct {
	ID        int64
	Name      string
	IsApplied bool
	AppliedAt time.Time
}

func GetStatus(db *sql.DB, opts ...OptionsFunc) ([]MigrationResult, error) {
	ctx := context.Background()
	return GetStatusContext(ctx, db, opts...)
}

func GetStatusContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) ([]MigrationResult, error) {
	option := applyOptions(opts)
	migrator, err := gomigrator.New(ctx, db, option.dialect, option.schema, option.table)
	if err != nil {
		return nil, err
	}

	migrations, err := getMigrations(option.migrationType, option.directory)
	if err != nil {
		return nil, err
	}

	results, err := migrator.GetStatus(ctx, migrations, db)
	if err != nil {
		return nil, err
	}

	migrationResults := make([]MigrationResult, 0, len(results))
	for _, r := range results {
		migrationResults = append(migrationResults, MigrationResult{
			ID:        r.ID,
			Name:      r.Name,
			IsApplied: !r.AppliedAt.IsZero(),
			AppliedAt: r.AppliedAt,
		})
	}

	return migrationResults, nil
}

func GetDBVersion(db *sql.DB, opts ...OptionsFunc) (int64, error) {
	ctx := context.Background()
	return GetDBVersionContext(ctx, db, opts...)
}

func GetDBVersionContext(ctx context.Context, db *sql.DB, opts ...OptionsFunc) (int64, error) {
	option := applyOptions(opts)
	migrator, err := gomigrator.New(ctx, db, option.dialect, option.schema, option.table)
	if err != nil {
		return -1, err
	}

	version, err := migrator.GetDBVersion(ctx, db)
	if err != nil {
		return -1, err
	}

	return version, nil
}

var ErrUnsupportedMigrationType = errors.New("migration type is unsupported")

func getMigrations(migrationType, directory string) (m gomigrator.Migrations, err error) {
	switch migrationType {
	case MigrationTypeGo:
		m, err = registerGoMigrations(globalGoMigrations)
	case MigrationTypeSQL:
		m, err = sqlmigration.SeekMigrations(directory, nil)
	default:
		return nil, ErrUnsupportedMigrationType
	}
	return m, err
}

func rollback(ctx context.Context, db *sql.DB, redo bool, opts []OptionsFunc) error {
	option := applyOptions(opts)
	migrator, err := gomigrator.New(ctx, db, option.dialect, option.schema, option.table)
	if err != nil {
		return err
	}

	migrations, err := getMigrations(option.migrationType, option.directory)
	if err != nil {
		return err
	}

	if err = migrator.Down(ctx, migrations, db, redo); err != nil {
		return err
	}

	return nil
}
