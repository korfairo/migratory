package command

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/korfairo/migratory/internal/gomigrator"
	"github.com/korfairo/migratory/internal/sqlmigration"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [--dir <path>] [-d <db-string>] [-s <schema>] [-t <table>]",
	Short: "Shows migration statuses",
	Long: `The "status" command shows table with migration statuses,
according existing migrations in your directory and in the database migrations table.
Command creates migrations table if not exists.`,
	Example: `migratory status -c /etc/config.yml
migratory status -d postgresql://role:password@127.0.0.1:5432/database --dir example/migrations/
migratory status -d postgresql://role:password@127.0.0.1:5432/database --dir migrations/ -t my_migrations_table`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := status(config.Dir, config.Schema, config.Table); err != nil {
			fmt.Printf("unable to get migrations status: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func status(dir, schema, table string) error {
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

	migrationStatuses, err := migrator.GetStatus(ctx, migrations, db)
	if err != nil {
		return fmt.Errorf("failed to GetStatus(...): %w", err)
	}

	printStatus(migrationStatuses)

	return nil
}

func printStatus(migrationStatuses []gomigrator.MigrationResult) {
	w := tabwriter.NewWriter(os.Stdout, 3, 1, 2, ' ', 0)

	_, err := fmt.Fprintf(w, "ID\tName\tApplied\tDate\t\n")
	if err != nil {
		fmt.Println("failed to print status string")
		os.Exit(1)
	}

	for _, ms := range migrationStatuses {
		_, err = fmt.Fprintf(w, "%d\t%s\t%t\t%v\t\n",
			ms.ID, ms.Name, !ms.AppliedAt.IsZero(), ms.AppliedAt.Format(time.DateTime))
		if err != nil {
			fmt.Println("failed to print status string")
			os.Exit(1)
		}
	}

	err = w.Flush()
	if err != nil {
		fmt.Println("failed to flush tabwriter")
	}
}
