package command

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/korfairo/migratory/internal/gomigrator"
	"github.com/spf13/cobra"
)

var dbVersionCmd = &cobra.Command{
	Use:   "dbversion [-d <db-string>] [-s <schema>] [-t <table>]",
	Short: "Shows the DB version (id of the last applied migration",
	Long: `The "dbversion" command prints the id of the last applied migration 
from migrations table in your database. Command creates migrations table if not exists.`,
	Example: `migratory dbversion -c /etc/config.yml
migratory dbversion -d postgresql://role:password@127.0.0.1:5432/database
migratory dbversion -d postgresql://role:password@127.0.0.1:5432/database -s my_schema -t my_migrations_table`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getDBVersion(config.Schema, config.Table)
		if err != nil {
			fmt.Printf("unable to get database version: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("database version: %d\n", version)
	},
}

func init() {
	rootCmd.AddCommand(dbVersionCmd)
}

func getDBVersion(schema, table string) (int64, error) {
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return 0, fmt.Errorf("could not open database: %w", err)
	}

	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("failed to close database connection")
		}
	}()

	ctx := context.Background()
	migrator, err := gomigrator.New(ctx, db, "postgres", schema, table)
	if err != nil {
		return 0, fmt.Errorf("could not create migrator: %w", err)
	}

	version, err := migrator.GetDBVersion(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("failed to GetDBVersion(...): %w", err)
	}

	return version, nil
}
