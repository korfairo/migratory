package sqlmigration

import (
	"testing"

	"github.com/korfairo/migratory/internal/require"
)

func TestParseMigrationFileName(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		fileName string
		wantID   int64
		wantName string
		wantErr  error
	}{
		"empty name":     {fileName: "1_.sql", wantID: 1, wantName: "", wantErr: nil},
		"simple":         {fileName: "1_name.sql", wantID: 1, wantName: "name", wantErr: nil},
		"underscores":    {fileName: "2_create_orders_table.sql", wantID: 2, wantName: "create_orders_table", wantErr: nil},
		"symbols":        {fileName: "3_@#$%^&*()-+=?><{}[].sql", wantID: 3, wantName: "@#$%^&*()-+=?><{}[]", wantErr: nil},
		"time version":   {fileName: "20240101_new.sql", wantID: 20240101, wantName: "new", wantErr: nil},
		"two dots":       {fileName: "20240102_new.old.sql", wantID: 20240102, wantName: "new.old", wantErr: nil},
		"zero":           {fileName: "01_new.sql", wantID: 1, wantName: "new", wantErr: nil},
		"multiple zeros": {fileName: "000001_new.sql", wantID: 1, wantName: "new", wantErr: nil},
		"with path":      {fileName: "./migrations/000001_new.sql", wantID: 1, wantName: "new", wantErr: nil},

		"empty string":  {fileName: "", wantID: -1, wantName: "", wantErr: ErrNoSeparator},
		"empty version": {fileName: "_name.sql", wantID: -1, wantName: "", wantErr: ErrParseID},
		"no underscore": {fileName: "2.sql", wantID: -1, wantName: "", wantErr: ErrNoSeparator},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotID, gotMigrationName, err := ParseMigrationFileName(test.fileName)
			require.ErrorIs(t, err, test.wantErr, "ParseMigrationFilePath(...): error")

			require.Int64(t, gotID, test.wantID, "ParseMigrationFilePath(...): migration ID")

			require.String(t, gotMigrationName, test.wantName, "ParseMigrationFilePath(...): migration name")
		})
	}
}
