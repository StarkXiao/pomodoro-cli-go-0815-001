package storage

import (
	"path/filepath"
	"testing"
	"time"

	"localpomodoro/internal/pomodoro"
)

func TestJSONStoreLoadMissingFileReturnsDefaults(t *testing.T) {
	store := JSONStore{Path: filepath.Join(t.TempDir(), "missing.json")}
	data, err := store.Load()
	if err != nil {
		t.Fatalf("expected missing file to return defaults, got %v", err)
	}
	if data.Config.FocusDuration != 25*time.Minute {
		t.Fatalf("expected default focus duration, got %s", data.Config.FocusDuration)
	}
}

func TestJSONStoreSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")
	store := JSONStore{Path: path}

	now := time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)
	original := pomodoro.DataFile{
		Config: pomodoro.DefaultConfig(),
		Runtime: pomodoro.RuntimeState{
			CompletedFocusSinceLongBreak: 2,
			UpdatedAt:                    now,
		},
		Sessions: []pomodoro.SessionRecord{
			{
				ID:           "one",
				TaskName:     "Write design doc",
				StartTime:    now,
				EndTime:      now.Add(25 * time.Minute),
				PlannedFocus: 25 * time.Minute,
				Status:       pomodoro.StatusCompleted,
				NextBreak:    pomodoro.BreakShort,
			},
		},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("expected load to succeed, got %v", err)
	}
	if loaded.Runtime.CompletedFocusSinceLongBreak != 2 {
		t.Fatalf("expected runtime counter 2, got %d", loaded.Runtime.CompletedFocusSinceLongBreak)
	}
	if len(loaded.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(loaded.Sessions))
	}
	if loaded.Sessions[0].TaskName != "Write design doc" {
		t.Fatalf("unexpected task name %q", loaded.Sessions[0].TaskName)
	}
}
