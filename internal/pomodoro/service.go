package pomodoro

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

type ProgressRunner func(ctx context.Context, duration time.Duration, out io.Writer) error

type Service struct {
	Now      func() time.Time
	Progress ProgressRunner
}

func NewService() Service {
	return Service{
		Now:      time.Now,
		Progress: RunProgress,
	}
}

func (s Service) RunFocus(ctx context.Context, out io.Writer, runtime RuntimeState, cfg Config, taskName string) (SessionRecord, RuntimeState, error) {
	if err := cfg.Validate(); err != nil {
		return SessionRecord{}, runtime, err
	}
	taskName = strings.TrimSpace(taskName)
	if taskName == "" {
		return SessionRecord{}, runtime, fmt.Errorf("task name is required")
	}

	nowFn := s.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	progressFn := s.Progress
	if progressFn == nil {
		progressFn = RunProgress
	}

	start := nowFn()
	record := SessionRecord{
		ID:           NewSessionID(start),
		TaskName:     taskName,
		StartTime:    start,
		PlannedFocus: cfg.FocusDuration,
		Status:       StatusInterrupted,
	}

	err := progressFn(ctx, cfg.FocusDuration, out)
	end := nowFn()
	record.EndTime = end
	if err != nil {
		if ctx.Err() != nil {
			record.Status = StatusInterrupted
			record.NextBreak = ""
			return record, runtime, ctx.Err()
		}
		return SessionRecord{}, runtime, err
	}

	record.Status = StatusCompleted
	runtime.CompletedFocusSinceLongBreak++
	if runtime.CompletedFocusSinceLongBreak > cfg.LongBreakEvery {
		record.NextBreak = BreakLong
		runtime.CompletedFocusSinceLongBreak = 0
	} else {
		record.NextBreak = BreakShort
	}
	runtime.UpdatedAt = end
	return record, runtime, nil
}
