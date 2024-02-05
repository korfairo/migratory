package config

import (
	"testing"

	"github.com/korfairo/migratory/internal/require"
)

func TestReadConfig(t *testing.T) {
	tests := map[string]struct {
		path string
		want *Config
		err  error
	}{
		"valid config": {
			path: "testdata/valid.yml",
			want: &Config{
				Dir:    "/path/to/directory",
				DSN:    "postgres://user:password@localhost:5432/my_db",
				Schema: "public",
				Table:  "migrations",
			},
			err: nil,
		},
		"empty config": {
			path: "testdata/empty.yml",
			want: &defaultConfig,
			err:  nil,
		},
		"nonexistent file": {
			path: "testdata/nonexistent.yml",
			want: nil,
			err:  ErrReadConfigFile,
		},
		"invalid config": {
			path: "testdata/invalid.yml",
			want: nil,
			err:  ErrUnmarshalFailure,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ReadConfig(test.path)

			require.Equal(t, got, test.want, "ReadConfig(...) config")

			require.ErrorIs(t, err, test.err, "ReadConfig(...)")
		})
	}
}
