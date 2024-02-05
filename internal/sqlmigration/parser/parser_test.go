package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/korfairo/migratory/internal/require"
)

const (
	validDataPath   = "testdata/valid/"
	invalidDataPath = "testdata/invalid/"
)

func TestEndsWithSemicolon(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		line string
		want bool
	}{
		"only semicolon":            {line: ";", want: true},
		"semicolon":                 {line: "END;", want: true},
		"with comment":              {line: "END; -- comment", want: true},
		"with spaces":               {line: "END   ; -- comment", want: true},
		"empty":                     {line: "", want: false},
		"no semicolon":              {line: "END", want: false},
		"no semicolon with comment": {line: "END -- comment", want: false},
		"semicolon in comment":      {line: "END -- comment ;", want: false},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := endsWithSemicolon(test.line)
			require.Bool(t, got, test.want, "endsWithSemicolon(...) result")
		})
	}
}

func TestParseValid(t *testing.T) {
	t.Parallel()
	fileNames := getDirectoryFilenames(t, validDataPath)
	for _, path := range fileNames {
		file := openFile(t, path)

		_, err := ParseMigration(file)
		require.NoError(t, err, fmt.Sprintf("ParseMigration(...) must execute without error, file %s", path))

		closeFile(t, file)
	}
}

func TestParseInvalid(t *testing.T) {
	t.Parallel()
	fileNames := getDirectoryFilenames(t, invalidDataPath)
	for _, path := range fileNames {
		file := openFile(t, path)

		_, err := ParseMigration(file)
		require.Error(t, err, fmt.Sprintf("ParseMigration(...) must execute with error, file %s", path))

		closeFile(t, file)
	}
}

func TestParseSplitStatements(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		sql       string
		upCount   int
		downCount int
	}{
		"complex": {
			sql:       complexMigration,
			upCount:   6,
			downCount: 4,
		},
		"noUp": {
			sql:       noUpMigration,
			upCount:   0,
			downCount: 1,
		},
		"noDown": {
			sql:       noDownMigration,
			upCount:   1,
			downCount: 0,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			migration, err := ParseMigration(strings.NewReader(test.sql))
			require.NoError(t, err, "ParseMigration(...)")
			require.Int(t, len(migration.UpStatements), test.upCount, "UpStatements count")
			require.Int(t, len(migration.DownStatements), test.downCount, "DownStatements count")
		})
	}
}

var complexMigration = `
-- +migrate up no_transaction
-- sql comment #1
CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    name VARCHAR(50),
    price DECIMAL(10,2),
    description TEXT
);

-- comment #2
-- +migrate statement_begin

-- comment #3
CREATE OR REPLACE FUNCTION add_product(name VARCHAR(50), price DECIMAL(10,2), description TEXT, category_id INTEGER)
RETURNS VOID AS $$
DECLARE
product_id INTEGER;
BEGIN
    INSERT INTO products (name, price, description) VALUES (name, price, description) RETURNING id INTO product_id;
    INSERT INTO product_categories (product_id, category_id) VALUES (product_id, category_id);
END;
$$ LANGUAGE plpgsql;

-- +migrate statement_end

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    hire_date DATE NOT NULL DEFAULT CURRENT_DATE,
    salary DECIMAL(10,2) NOT NULL
);

CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    manager_id INTEGER REFERENCES employees(id) ON DELETE SET NULL

);

ALTER TABLE employees ADD COLUMN created_at TIMESTAMP DEFAULT NOW();

CREATE INDEX idx_departments_manager_id ON departments(manager_id);

-- +migrate down
DROP FUNCTION add_product(VARCHAR(50), DECIMAL(10,2), TEXT, INTEGER);

DROP TABLE products;

DROP TABLE departments
;
DROP TABLE employees;
`

var noDownMigration = `
-- +migrate up
-- no time to add down statement, will do later
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL
);
`

var noUpMigration = `
-- finally added down statement
-- +migrate down
DROP TABLE users;
`

func getDirectoryFilenames(t *testing.T, path string) []string {
	t.Helper()
	fileNames, err := filepath.Glob(path + "*.sql")
	require.NoError(t, err, fmt.Sprintf("failed to get file names at path %s", path))
	return fileNames
}

func openFile(t *testing.T, path string) *os.File {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err, fmt.Sprintf("failed to open file at path %s", path))
	return file
}

func closeFile(t *testing.T, file *os.File) {
	t.Helper()
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close file")
	}
}
