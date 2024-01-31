package sqlmigration

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/korfairo/migratory/internal/require"
	"github.com/korfairo/migratory/internal/sqlmigration/testdata/mock"
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

		"empty version": {fileName: "_name.sql", wantID: 0, wantName: "", wantErr: ErrParseID},
		"no underscore": {fileName: "2.sql", wantID: 0, wantName: "", wantErr: ErrNoSeparator},
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

func TestSeekMigrations(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		directory string
		prepare   func(seeker *mock.MockFileSystem)
		wantLen   int
		wantErr   error
	}{
		"incorrect path": {
			directory: "incorrect/path/",
			prepare: func(seeker *mock.MockFileSystem) {
				seeker.EXPECT().Stat("incorrect/path/").Return(nil, os.ErrNotExist)
			},
			wantLen: 0,
			wantErr: ErrDirectoryCheck,
		},
		"glob error": {
			directory: "correct/path/",
			prepare: func(seeker *mock.MockFileSystem) {
				seeker.EXPECT().Stat("correct/path/").Return(nil, nil)
				seeker.EXPECT().Glob("correct/path/*.sql").Return(nil, os.ErrPermission)
			},
			wantLen: 0,
			wantErr: ErrGlobMigrations,
		},
		"no migrations": {
			directory: "correct/path/",
			prepare: func(seeker *mock.MockFileSystem) {
				seeker.EXPECT().Stat("correct/path/").Return(nil, nil)
				seeker.EXPECT().Glob("correct/path/*.sql").Return([]string{}, nil)
			},
			wantLen: 0,
			wantErr: ErrNoMigrationFiles,
		},
		"no separator": {
			directory: "correct/path/",
			prepare: func(seeker *mock.MockFileSystem) {
				seeker.EXPECT().Stat("correct/path/").Return(nil, nil)
				fileNames := []string{"01name.sql"}
				seeker.EXPECT().Glob("correct/path/*.sql").Return(fileNames, nil)
			},
			wantLen: 0,
			wantErr: ErrNoSeparator,
		},
		"incorrect id": {
			directory: "correct/path/",
			prepare: func(seeker *mock.MockFileSystem) {
				seeker.EXPECT().Stat("correct/path/").Return(nil, nil)
				fileNames := []string{"0.1_name.sql"}
				seeker.EXPECT().Glob("correct/path/*.sql").Return(fileNames, nil)
			},
			wantLen: 0,
			wantErr: ErrParseID,
		},
		"valid names": {
			directory: "correct/path/",
			prepare: func(seeker *mock.MockFileSystem) {
				seeker.EXPECT().Stat("correct/path/").Return(nil, nil)
				fileNames := []string{
					"1_create_table.sql",
					"2_add_index.sql",
					"3_new_orders.sql",
					"4_add_column.sql",
					"5_drop_index.sql",
				}
				seeker.EXPECT().Glob("correct/path/*.sql").Return(fileNames, nil)
			},
			wantLen: 5,
			wantErr: nil,
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			seeker := mock.NewMockFileSystem(ctrl)

			if test.prepare != nil {
				test.prepare(seeker)
			}

			gotMigrations, err := SeekMigrations(test.directory, seeker)
			require.ErrorIs(t, err, test.wantErr, "SeekMigrations(...) error")
			require.Int(t, len(gotMigrations), test.wantLen, "SeekMigrations(...) migrations count")
		})
	}
}
