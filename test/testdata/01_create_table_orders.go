package gomigrations

import (
	"context"
	"database/sql"

	"github.com/korfairo/migratory"
)

func init() {
	migratory.AddMigration(up01, down01)
}

func up01(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "CREATE TABLE orders (id INTEGER)")
	return err
}

func down01(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE orders")
	return err
}
