package sqlmigration

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/korfairo/migratory/internal/gomigrator"
)

var (
	ErrNoMigrationFiles = errors.New("migration files *.sql not found")
	ErrDuplicatedID     = errors.New("duplicated migrations ID detected")
	ErrGlobMigrations   = errors.New("unable to search migrations in directory")
	ErrDirectoryCheck   = errors.New("unable to check directory existence")
)

type FileSystem interface {
	Open(name string) (fs.File, error)
	Stat(name string) (os.FileInfo, error)
	Glob(pattern string) ([]string, error)
}

// SeekMigrations searches for .sql migration files in directory.
//
// It parses file names and returns gomigrator.Migrations sorted by version ascending.
func SeekMigrations(dir string, fs FileSystem) (gomigrator.Migrations, error) {
	var migrations gomigrator.Migrations
	if fs == nil {
		fs = osWrapper{}
	}

	if _, err := fs.Stat(dir); err != nil {
		return nil, errors.Join(ErrDirectoryCheck, err)
	}

	sqlMigrationFileNames, err := fs.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, errors.Join(ErrGlobMigrations, err)
	}

	migrationFilesCount := len(sqlMigrationFileNames)
	if migrationFilesCount == 0 {
		return nil, ErrNoMigrationFiles
	}

	uniqueIDMap := make(map[int64]struct{}, migrationFilesCount)
	for _, filePath := range sqlMigrationFileNames {
		id, name, err := ParseMigrationFileName(filePath)
		if err != nil {
			return nil, fmt.Errorf("file name %s doesn't match the template {id}_{name}.sql: %w", filePath, err)
		}

		if _, exist := uniqueIDMap[id]; exist {
			return nil, fmt.Errorf("ID %d, %w", id, ErrDuplicatedID)
		}
		uniqueIDMap[id] = struct{}{}

		migrations = append(migrations,
			gomigrator.NewMigrationWithPreparer(id, name, newSQLPreparer(filePath, fs)))
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID() < migrations[j].ID()
	})

	return migrations, nil
}

const FileNameSeparator = "_"

var (
	ErrNoSeparator = errors.New("no separator found in file name")
	ErrParseID     = errors.New("unable to parse ID in file name")
)

func ParseMigrationFileName(fileName string) (id int64, migrationName string, err error) {
	base := filepath.Base(fileName)
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(fileName))

	separatorIdx := strings.Index(nameWithoutExt, FileNameSeparator)
	if separatorIdx < 0 {
		return -1, "", ErrNoSeparator
	}

	id, err = strconv.ParseInt(nameWithoutExt[:separatorIdx], 10, 64)
	if err != nil {
		return -1, "", ErrParseID
	}

	migrationName = nameWithoutExt[separatorIdx+1:]

	return id, migrationName, nil
}
