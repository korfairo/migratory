package gomigrator

import (
	"testing"

	"github.com/korfairo/migratory/internal/gomigrator/dialect"
	"github.com/korfairo/migratory/internal/require"
)

func TestNewStore(t *testing.T) {
	type args struct {
		dbDialect  string
		schemaName string
		tableName  string
	}
	tests := map[string]struct {
		args    args
		want    Store
		wantErr error
	}{
		"existing dialect": {
			args: args{
				dbDialect:  DialectPostgres,
				schemaName: "public",
				tableName:  "migrations",
			}, want: migrationStore{
				"public",
				"migrations",
				&dialect.Postgres{},
			},
			wantErr: nil,
		},
		"unknown dialect": {
			args: args{
				dbDialect:  "mysql",
				schemaName: "public",
				tableName:  "migrations",
			}, want: nil,
			wantErr: ErrUnsupportedDialect,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := newStore(test.args.dbDialect, test.args.schemaName, test.args.tableName)
			require.ErrorIs(t, err, test.wantErr, "newStore(...) error")
			require.Equal(t, got, test.want, "newStore(...) new store")
		})
	}
}
