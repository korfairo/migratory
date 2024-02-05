package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Dir    string `yaml:"directory"`
	DSN    string `yaml:"dsn"`
	Schema string `yaml:"schema"`
	Table  string `yaml:"table"`
}

var defaultConfig = Config{
	Dir:    ".",
	DSN:    "",
	Schema: "public",
	Table:  "migrations",
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
	if cfg.Dir == "" {
		cfg.Dir = defaultConfig.Dir
	}

	if cfg.Schema == "" {
		cfg.Schema = defaultConfig.Schema
	}

	if cfg.Table == "" {
		cfg.Table = defaultConfig.Table
	}
}

func expandConfig(cfg *Config) {
	cfg.Dir = os.ExpandEnv(cfg.Dir)
	cfg.DSN = os.ExpandEnv(cfg.DSN)
	cfg.Schema = os.ExpandEnv(cfg.Schema)
	cfg.Table = os.ExpandEnv(cfg.Table)
}
