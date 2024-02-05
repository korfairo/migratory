package sqlmigration

import (
	"io/fs"
	"os"
	"path/filepath"
)

type osWrapper struct{}

func (osWrapper) Open(name string) (fs.File, error) { return os.Open(name) }

func (osWrapper) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

func (osWrapper) Glob(pattern string) ([]string, error) { return filepath.Glob(pattern) }
