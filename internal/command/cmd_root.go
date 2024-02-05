package command

import (
	"fmt"
	"os"

	cfg "github.com/korfairo/migratory/internal/config"
	"github.com/spf13/cobra"
)

var (
	configPath string
	config     cfg.Config
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "migratory",
	Short: "Migratory is a migration tool and go package",
	Long: `Migratory is a migration tool and go package designed to simplify the 
process of migrating data and schema between databases. Now it supports PostgreSQL 
database dialect and allows you to choose between SQL or Go migration formats. 
With Migratory, you can easily create, apply, and roll back migrations using simple CLI commands. 
To get started, simply refer to the README.md file or write command "help" for more information 
on how to use this tool.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "",
		"path to yaml config")

	rootCmd.PersistentFlags().StringVar(&config.Dir, "dir", ".",
		"directory with .sql migration files")
	rootCmd.PersistentFlags().StringVarP(&config.DSN, "db", "d", "",
		"database connection string")
	rootCmd.PersistentFlags().StringVarP(&config.Schema, "schema", "s", "public",
		"name of database schema with migrations table")
	rootCmd.PersistentFlags().StringVarP(&config.Table, "table", "t", "migrations",
		"name of migrations table")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if configPath == "" {
		return
	}

	newConfig, err := cfg.ReadConfig(configPath)
	if err != nil {
		fmt.Printf("failed to read config: %s\n", err)
		os.Exit(1)
	}

	config = *newConfig
}
