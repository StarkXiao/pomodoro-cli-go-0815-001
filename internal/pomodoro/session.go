package pomodoro

import (
	"fmt"
	"strings"
	"time"
)

func (s SessionRecord) Validate() error {
	if strings.TrimSpace(s.TaskName) == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if s.StartTime.IsZero() {
		return fmt.Errorf("start time cannot be empty")
	}
	if s.EndTime.IsZero() {
		return fmt.Errorf("end time cannot be empty")
	}
	if s.EndTime.Before(s.StartTime) {
		return fmt.Errorf("end time cannot be before start time")
	}
	if s.PlannedFocus <= 0 {
		return fmt.Errorf("planned focus must be greater than zero")
	}
	switch s.Status {
	case StatusCompleted, StatusInterrupted:
	default:
		return fmt.Errorf("unknown status %q", s.Status)
	}
	if s.NextBreak != "" && s.NextBreak != BreakShort && s.NextBreak != BreakLong {
		return fmt.Errorf("unknown break type %q", s.NextBreak)
	}
	return nil
}

func (s SessionRecord) toJSON() sessionJSON {
	return sessionJSON{
		ID:           s.ID,
		TaskName:     s.TaskName,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
		PlannedFocus: s.PlannedFocus.String(),
		Status:       s.Status,
		NextBreak:    s.NextBreak,
	}
}

func sessionFromJSON(in sessionJSON) (SessionRecord, error) {
	duration, err := ParseDurationInput(in.PlannedFocus)
	if err != nil {
		return SessionRecord{}, fmt.Errorf("parse planned focus: %w", err)
	}
	record := SessionRecord{
		ID:           in.ID,
		TaskName:     in.TaskName,
		StartTime:    in.StartTime,
		EndTime:      in.EndTime,
		PlannedFocus: duration,
		Status:       in.Status,
		NextBreak:    in.NextBreak,
	}
	return record, record.Validate()
}

func NewSessionID(at time.Time) string {
	return at.Format("20060102T150405.000000000")
}
