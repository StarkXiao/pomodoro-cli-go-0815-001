package storage

import (
	"errors"
	"fmt"
	"os"
)

var ErrDataLocked = errors.New("data file is locked by another running command")

type FileLock struct {
	path string
}

func AcquireFileLock(path string) (*FileLock, error) {
	if path == "" {
		return nil, fmt.Errorf("lock path is required")
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, ErrDataLocked
		}
		return nil, fmt.Errorf("create lock file: %w", err)
	}

	pidText := fmt.Sprintf("%d\n", os.Getpid())
	if _, err := file.WriteString(pidText); err != nil {
		file.Close()
		os.Remove(path)
		return nil, fmt.Errorf("write lock file: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("close lock file: %w", err)
	}

	return &FileLock{path: path}, nil
}

func (l *FileLock) Release() error {
	if l == nil || l.path == "" {
		return nil
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove lock file %s: %w", l.path, err)
	}
	return nil
}
