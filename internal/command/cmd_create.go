package command

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create {<name>} {sql|go} [--dir <path>]",
	Short: "Creates .sql or .go migration template",
	Long: `This command creates .sql or .go file with standard migration template. 
Default directory is your current one, pass arg with -d flag to choose another.
Name of the file matches the format {id}_{name}.sql, where id is a unique number of migration.
The command writes current UTC time as a migration id, for example: 20060102150405_name.sql`,
	Example: `migratory create my_migration go
migratory create my_migration sql
migratory create my_migration sql -d ./example/migrations`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := create(config.Dir, args[0], args[1]); err != nil {
			fmt.Printf("unable to create template: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

const timeNumberFormat = "20060102150405"

func create(dir, name, migrationType string) error {
	var template *[]byte
	switch migrationType {
	case "go":
		template = &templateGo
	case "sql":
		template = &templateSQL
	default:
		return fmt.Errorf("unsupported migration type %s", migrationType)
	}

	id := time.Now().UTC().Format(timeNumberFormat)
	path := fmt.Sprintf("%s/%s_%s.%s", dir, id, name, migrationType)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Write(*template)
	if err != nil {
		return fmt.Errorf("failed to write template, file %s, type %s: %w", path, migrationType, err)
	}
	return nil
}

var templateSQL = []byte(`-- +migrate up
-- +migrate statement_begin
SELECT 'up SQL query';
-- +migrate statement_end

-- +migrate down
-- +migrate statement_begin
SELECT 'down SQL query';
-- +migrate statement_end
`)

var templateGo = []byte(`package migrations

import (
	"context"
	"database/sql"

	"github.com/korfairo/migratory"
)

func init() {
	migratory.AddMigration(up01, down01)
}

func up01(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func down01(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`)
