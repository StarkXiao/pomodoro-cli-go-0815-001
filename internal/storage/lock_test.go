package storage

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestAcquireFileLockPreventsConcurrentWriter(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "pomodoro.lock")

	first, err := AcquireFileLock(lockPath)
	if err != nil {
		t.Fatalf("expected first lock to succeed, got %v", err)
	}
	defer first.Release()

	second, err := AcquireFileLock(lockPath)
	if !errors.Is(err, ErrDataLocked) {
		if second != nil {
			second.Release()
		}
		t.Fatalf("expected second lock attempt to fail with ErrDataLocked, got %v", err)
	}
}
