package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateDirectoryIfNotExists(path string) error {
	dir, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}

	if dir.IsDir() {
		return nil
	}

	return fmt.Errorf("path is not a directory: %s", path)
}

func CreateFile(dir string, filename string) (*os.File, error) {
	err := CreateDirectoryIfNotExists(dir)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, filename)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return os.Create(path)
	}
	return nil, fmt.Errorf("file already exists: %s", path)
}
