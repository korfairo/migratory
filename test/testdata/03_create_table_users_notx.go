package gomigrations

import (
	"context"
	"database/sql"

	"github.com/korfairo/migratory"
)

func init() {
	migratory.AddMigrationNoTx(up03, down03)
}

func up03(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER)")
	return err
}

func down03(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "DROP TABLE users")
	return err
}
