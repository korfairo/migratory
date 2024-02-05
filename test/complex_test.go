//go:build integration
// +build integration

package test

import (
	"database/sql"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/korfairo/migratory"
	"github.com/korfairo/migratory/internal/gomigrator"
	"github.com/korfairo/migratory/internal/require"
	_ "github.com/korfairo/migratory/test/testdata"
	_ "github.com/lib/pq"
)

// dsn - connection string to test DB.
// Do not use production DB, public schema will be dropped!
var dsn = flag.String("dsn", "", "")

var migrationsTable = "go_migrations"

const (
	lastTestMigrationID int64 = 3
	testMigrationsCount int   = 3
)

func TestGoMigrations(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() {
		resetPublicSchema(t, db)
		err := db.Close()
		require.NoError(t, err, "db.Close()")
	})

	// number of applied migrations must be equal to testMigrationsCount
	appliedCount, err := migratory.Up(db, migratory.WithTable(migrationsTable))
	require.NoError(t, err, "migratory.Up(...) error")
	require.Int(t, appliedCount, testMigrationsCount, "migratory.Up(...) applied migrations count")

	// checking number of created tables in public schema (including migrations table)
	tableCount := getTableCount(t, db)
	if tableCount == 0 {
		t.Fatalf("no tables have been created in the database")
	}

	// checking migratory default options
	migratory.SetTable(migrationsTable)

	// checking db version
	dbVersion, err := migratory.GetDBVersion(db)
	require.NoError(t, err, "migratory.GetDBVersion(...) error")
	require.Int64(t, dbVersion, lastTestMigrationID, "dbVersion doesn't match the last migration ID")

	// checking status and number of rows in migrations table
	results, err := migratory.GetStatus(db)
	require.NoError(t, err, "migratory.GetStatus(...) error")

	migrationsCount := getMigrationsTableRowCount(t, db)
	require.Int(t, len(results), migrationsCount,
		"migratory.GetStatus(...) results must be equal to migrationsCount")

	// ensure all migrations are applied and maxID is equal to lastTestMigrationID
	var maxID int64
	var appliedAt time.Time
	for _, result := range results {
		require.Bool(t, result.IsApplied, true, "all migrations must be applied")

		if maxID < result.ID {
			maxID = result.ID
			appliedAt = result.AppliedAt
		}
	}

	require.Int64(t, maxID, dbVersion, "incorrect max ID in migration results")
	require.Bool(t, appliedAt.IsZero(), false, "last migration time must not be zero")

	dbMaxID, dbAppliedAt := getLastMigrationResult(t, db)
	require.Int64(t, maxID, dbMaxID,
		"migratory.GetStatus() maxID is not equal to the last migration ID got from DB")
	require.Time(t, appliedAt, dbAppliedAt,
		"migratory.GetStatus() appliedAt is not equal to the last migration applied_at got from DB")

	// ensure that AppliedAt has changed after migratory.Redo(...)
	err = migratory.Redo(db)
	require.NoError(t, err, "migratory.Redo(...) error")

	newDBMaxID, newDBAppliedAt := getLastMigrationResult(t, db)
	require.Int64(t, newDBMaxID, dbMaxID, "last migration ID must not change after migratory.Redo()")

	if !newDBAppliedAt.After(dbAppliedAt) {
		t.Fatalf("applied_at of last migration must increase after migratory.Redo(...)")
	}

	// checking migrations roll back with migratory.Down(db)
	for i := 1; i <= testMigrationsCount; i++ {
		err = migratory.Down(db)
		require.NoError(t, err, "migratory.Down(...) error")

		migrationsCount = getMigrationsTableRowCount(t, db)
		require.Int(t, migrationsCount, testMigrationsCount-i,
			"migratory.Down(...) migrationsCount must decrease during rollbacks")
	}

	err = migratory.Down(db)
	require.ErrorIs(t, err, gomigrator.ErrNothingToRollback, "migratory.Down(...) error")
}

func getLastMigrationResult(t *testing.T, db *sql.DB) (id int64, at time.Time) {
	t.Helper()
	q := "SELECT id, applied_at FROM %s ORDER BY id DESC LIMIT 1;"
	rows, err := db.Query(fmt.Sprintf(q, migrationsTable))
	require.NoError(t, err, "db.Query(...) last id error")
	require.NoError(t, rows.Err(), "rows.Err() error")
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &at)
		require.NoError(t, err, "rows.Scan() last id error")
	}
	return
}

func getMigrationsTableRowCount(t *testing.T, db *sql.DB) (n int) {
	t.Helper()
	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(1) FROM %s;", migrationsTable))
	require.NoError(t, row.Scan(&n), "db.QueryRow(...) migrations count error")
	return
}

func getTableCount(t *testing.T, db *sql.DB) (n int) {
	t.Helper()
	q := "SELECT COUNT(table_name) FROM information_schema.tables WHERE table_schema = 'public';"
	row := db.QueryRow(q)
	require.NoError(t, row.Scan(&n), "db.QueryRow(...) table count error")
	return
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if *dsn == "" {
		t.Fatalf("database dsn is empty, pass it with flag -dsn")
	}

	db, err := sql.Open("postgres", *dsn)
	require.NoError(t, err, "sql.Open(...) error")

	err = db.Ping()
	require.NoError(t, err, "db.Ping() error")

	resetPublicSchema(t, db)

	return db
}

func resetPublicSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("DROP SCHEMA IF EXISTS public CASCADE;")
	require.NoError(t, err, "db.Exec(...) drop schema")

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS public;")
	require.NoError(t, err, "db.Exec(...) create schema")
}
