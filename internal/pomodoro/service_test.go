package pomodoro

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

func TestRunFocusUsesLongBreakWhenThresholdIsReached(t *testing.T) {
	cfg := Config{
		FocusDuration:      25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakEvery:     2,
	}

	nowValues := []time.Time{
		time.Date(2026, 8, 16, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 16, 9, 25, 0, 0, time.UTC),
	}
	callIndex := 0
	service := Service{
		Now: func() time.Time {
			value := nowValues[callIndex]
			callIndex++
			return value
		},
		Progress: func(ctx context.Context, duration time.Duration, out io.Writer) error {
			return nil
		},
	}

	record, runtime, err := service.RunFocus(context.Background(), &bytes.Buffer{}, RuntimeState{
		CompletedFocusSinceLongBreak: 1,
	}, cfg, "Write tests")
	if err != nil {
		t.Fatalf("expected run focus to succeed, got %v", err)
	}
	if record.Status != StatusCompleted {
		t.Fatalf("expected completed status, got %q", record.Status)
	}
	if record.NextBreak != BreakLong {
		t.Fatalf("expected next break %q, got %q", BreakLong, record.NextBreak)
	}
	if runtime.CompletedFocusSinceLongBreak != 0 {
		t.Fatalf("expected counter reset after long break, got %d", runtime.CompletedFocusSinceLongBreak)
	}
}
