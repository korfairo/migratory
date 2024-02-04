package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down [--dir <path>] [-d <db-string>] [-s <schema>] [-t <table>]",
	Short: "Rollback last applied migration",
	Long: `The "down" command rolls back the last applied migration.
Command creates migrations table if not exists.`,
	Example: `migratory down -c /etc/config.yml
migratory down -d postgresql://role:password@127.0.0.1:5432/database --dir example/migrations/
migratory down -d postgresql://role:password@127.0.0.1:5432/database --dir migrations/ -t my_migrations_table`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := rollback(config.Dir, config.Schema, config.Table, false); err != nil {
			fmt.Printf("unable to rollback migration: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("success: migration rolled back")
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
