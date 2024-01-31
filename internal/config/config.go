package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Directory string `yaml:"directory"`
	DBString  string `yaml:"db_string"`
	Schema    string `yaml:"db_schema"`
	Table     string `yaml:"db_table"`
}

var defaultConfig = Config{
	Directory: ".",
	DBString:  "",
	Schema:    "public",
	Table:     "migrations",
}

var (
	ErrReadConfigFile   = errors.New("failed to read config file")
	ErrUnmarshalFailure = errors.New("failed to unmarshal config")
)

func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadConfigFile, err)
	}

	cfg := &Config{}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnmarshalFailure, err)
	}

	setDefaultValues(cfg)

	expandConfig(cfg)

	return cfg, nil
}

func setDefaultValues(cfg *Config) {
	if cfg.Directory == "" {
		cfg.Directory = defaultConfig.Directory
	}

	if cfg.Schema == "" {
		cfg.Schema = defaultConfig.Schema
	}

	if cfg.Table == "" {
		cfg.Table = defaultConfig.Table
	}
}

func expandConfig(cfg *Config) {
	cfg.Directory = os.ExpandEnv(cfg.Directory)
	cfg.DBString = os.ExpandEnv(cfg.DBString)
	cfg.Schema = os.ExpandEnv(cfg.Schema)
	cfg.Table = os.ExpandEnv(cfg.Table)
}
