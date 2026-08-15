package report

import (
	"strings"
	"testing"
	"time"

	"localpomodoro/internal/pomodoro"
)

func TestBuildDailySummary(t *testing.T) {
	loc := time.UTC
	records := []pomodoro.SessionRecord{
		{
			ID:           "1",
			TaskName:     "Implement CLI",
			StartTime:    time.Date(2026, 8, 15, 9, 0, 0, 0, loc),
			EndTime:      time.Date(2026, 8, 15, 9, 25, 0, 0, loc),
			PlannedFocus: 25 * time.Minute,
			Status:       pomodoro.StatusCompleted,
		},
		{
			ID:           "2",
			TaskName:     "Review logs",
			StartTime:    time.Date(2026, 8, 15, 10, 0, 0, 0, loc),
			EndTime:      time.Date(2026, 8, 15, 10, 5, 0, 0, loc),
			PlannedFocus: 25 * time.Minute,
			Status:       pomodoro.StatusInterrupted,
		},
	}

	summary := BuildDailySummary(records, time.Date(2026, 8, 15, 0, 0, 0, 0, loc), loc)
	if summary.TotalSessions != 2 {
		t.Fatalf("expected 2 sessions, got %d", summary.TotalSessions)
	}
	if summary.CompletedFocus != 25*time.Minute {
		t.Fatalf("expected 25 minutes completed focus, got %s", summary.CompletedFocus)
	}
	if len(summary.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(summary.Tasks))
	}
}

func TestBuildWeeklyReport(t *testing.T) {
	loc := time.UTC
	records := []pomodoro.SessionRecord{
		{
			ID:           "1",
			TaskName:     "A",
			StartTime:    time.Date(2026, 8, 10, 9, 0, 0, 0, loc),
			EndTime:      time.Date(2026, 8, 10, 9, 25, 0, 0, loc),
			PlannedFocus: 25 * time.Minute,
			Status:       pomodoro.StatusCompleted,
		},
		{
			ID:           "2",
			TaskName:     "B",
			StartTime:    time.Date(2026, 8, 12, 9, 0, 0, 0, loc),
			EndTime:      time.Date(2026, 8, 12, 9, 5, 0, 0, loc),
			PlannedFocus: 25 * time.Minute,
			Status:       pomodoro.StatusInterrupted,
		},
	}

	weekly := BuildWeeklyReport(records, time.Date(2026, 8, 15, 0, 0, 0, 0, loc), loc)
	if weekly.CompletedSessions != 1 {
		t.Fatalf("expected 1 completed session, got %d", weekly.CompletedSessions)
	}
	if weekly.InterruptedSessions != 1 {
		t.Fatalf("expected 1 interrupted session, got %d", weekly.InterruptedSessions)
	}
	if len(weekly.Days) != 7 {
		t.Fatalf("expected 7 day rows, got %d", len(weekly.Days))
	}
	output := FormatWeeklyReport(weekly)
	if !strings.Contains(output, "2026-08-10") {
		t.Fatalf("expected weekly output to include Monday, got %q", output)
	}
}

func TestBuildReportsWithFixedZoneSessionTimes(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	records := []pomodoro.SessionRecord{
		{
			ID:           "1",
			TaskName:     "Demo",
			StartTime:    time.Date(2026, 8, 15, 19, 58, 59, 0, loc),
			EndTime:      time.Date(2026, 8, 15, 19, 59, 0, 0, loc),
			PlannedFocus: time.Second,
			Status:       pomodoro.StatusCompleted,
		},
	}

	summary := BuildDailySummary(records, time.Date(2026, 8, 15, 0, 0, 0, 0, time.Local), time.Local)
	if summary.TotalSessions != 1 {
		t.Fatalf("expected 1 daily session, got %d", summary.TotalSessions)
	}

	weekly := BuildWeeklyReport(records, time.Date(2026, 8, 15, 0, 0, 0, 0, time.Local), time.Local)
	if weekly.TotalSessions != 1 {
		t.Fatalf("expected 1 weekly session, got %d", weekly.TotalSessions)
	}
}

func TestBuildWeeklyReportUsesMondayAsWeekStart(t *testing.T) {
	loc := time.UTC
	weekly := BuildWeeklyReport(nil, time.Date(2026, 8, 15, 0, 0, 0, 0, loc), loc)
	if weekly.WeekStart != "2026-08-10" {
		t.Fatalf("expected week start 2026-08-10, got %s", weekly.WeekStart)
	}
	if weekly.WeekEnd != "2026-08-16" {
		t.Fatalf("expected week end 2026-08-16, got %s", weekly.WeekEnd)
	}
}
