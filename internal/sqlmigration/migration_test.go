package sqlmigration

import (
	"os"
	"sync"
	"testing"

	"github.com/korfairo/migratory/internal/gomigrator"
	"github.com/korfairo/migratory/internal/require"
)

func TestSQLPreparerPrepare(t *testing.T) {
	testFiles := &tmpFiles{}
	defer testFiles.RemoveAll(t)

	type fields struct {
		sourcePath string
	}

	tests := map[string]struct {
		fields         fields
		migrationData  []byte
		createTestFile func(t *testing.T, files *tmpFiles)
		want           *gomigrator.ExecutorContainer
		wantErr        bool
	}{
		"opening error": {
			fields: fields{
				sourcePath: "incorrect/path",
			},
			want:    nil,
			wantErr: true,
		},
		"parseMigration error, empty file": {
			fields: fields{
				sourcePath: "01_tmp_migration.sql",
			},
			createTestFile: func(t *testing.T, files *tmpFiles) {
				t.Helper()
				data := ""
				files.Create(t, "01_tmp_migration.sql", data)
			},
			want:    nil,
			wantErr: true,
		},
		"valid file": {
			fields: fields{
				sourcePath: "02_tmp_migration.sql",
			},
			createTestFile: func(t *testing.T, files *tmpFiles) {
				t.Helper()
				data := "-- +migrate up\n" +
					"SELECT COUNT(1);\n" +
					"-- +migrate down\n" +
					"SELECT COUNT(2);"
				files.Create(t, "02_tmp_migration.sql", data)
			},
			want: gomigrator.NewExecutorContainer(
				newSQLExecutor(
					[]string{"SELECT COUNT(1);\n"},
					[]string{"SELECT COUNT(2);\n"},
				),
			),
			wantErr: false,
		},
		"valid file no transaction": {
			fields: fields{
				sourcePath: "03_tmp_migration.sql",
			},
			createTestFile: func(t *testing.T, files *tmpFiles) {
				t.Helper()
				data := "-- +migrate up no_transaction\n" +
					"SELECT COUNT(1);\n" +
					"-- +migrate down\n" +
					"SELECT COUNT(2);"
				files.Create(t, "03_tmp_migration.sql", data)
			},
			want: gomigrator.NewExecutorContainerNoTx(
				newSQLExecutorNoTx(
					[]string{"SELECT COUNT(1);\n"},
					[]string{"SELECT COUNT(2);\n"},
				),
			),
			wantErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			preparer := newSQLPreparer(test.fields.sourcePath, osWrapper{})

			if test.createTestFile != nil {
				test.createTestFile(t, testFiles)
			}

			got, err := preparer.Prepare()

			if test.wantErr {
				require.Error(t, err, "SeekMigrations(...) error")
			} else {
				require.NoError(t, err, "SeekMigrations(...) error")
			}

			require.Equal(t, got, test.want, "SeekMigrations(...) ExecutorContainer")
		})
	}
}

type tmpFiles struct {
	mu        sync.Mutex
	fileNames []string
}

func (tf *tmpFiles) Create(t *testing.T, path string, data string) *os.File {
	t.Helper()
	tf.mu.Lock()
	defer tf.mu.Unlock()

	file, err := os.Create(path)
	require.NoError(t, err, "os.Create(...) error")

	_, err = file.WriteString(data)
	require.NoError(t, err, "file.WriteString(...) error")

	tf.fileNames = append(tf.fileNames, path)

	return file
}

func (tf *tmpFiles) RemoveAll(t *testing.T) {
	t.Helper()
	tf.mu.Lock()
	defer tf.mu.Unlock()

	for _, path := range tf.fileNames {
		err := os.Remove(path)
		require.NoError(t, err, "os.Remove(...) error")
	}

	tf.fileNames = []string{}
}
