package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"localpomodoro/internal/pomodoro"
)

type DailySummary struct {
	Date                string
	TotalSessions       int
	CompletedSessions   int
	InterruptedSessions int
	CompletedFocus      time.Duration
	Tasks               []string
}

type WeeklyDaySummary struct {
	Date                string
	CompletedSessions   int
	InterruptedSessions int
	CompletedFocus      time.Duration
}

type WeeklyReport struct {
	WeekStart           string
	WeekEnd             string
	TotalSessions       int
	CompletedSessions   int
	InterruptedSessions int
	CompletedFocus      time.Duration
	Days                []WeeklyDaySummary
}

func BuildDailySummary(records []pomodoro.SessionRecord, day time.Time, loc *time.Location) DailySummary {
	if loc == nil {
		loc = time.Local
	}
	target := normalizeDate(day.In(loc))
	targetKey := dateKey(target)
	taskSet := map[string]struct{}{}
	summary := DailySummary{Date: target.Format("2006-01-02")}

	for _, record := range records {
		start := record.StartTime.In(loc)
		if dateKey(start) != targetKey {
			continue
		}
		summary.TotalSessions++
		taskSet[record.TaskName] = struct{}{}
		summary.CompletedFocus += record.PlannedFocus
		switch record.Status {
		case pomodoro.StatusCompleted:
			summary.CompletedSessions++
		case pomodoro.StatusInterrupted:
			summary.InterruptedSessions++
		}
	}

	for task := range taskSet {
		summary.Tasks = append(summary.Tasks, task)
	}
	sort.Strings(summary.Tasks)
	return summary
}

func BuildWeeklyReport(records []pomodoro.SessionRecord, day time.Time, loc *time.Location) WeeklyReport {
	if loc == nil {
		loc = time.Local
	}
	base := day.In(loc)
	weekStart := startOfWeek(base)
	report := WeeklyReport{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekStart.AddDate(0, 0, 6).Format("2006-01-02"),
	}

	days := make([]WeeklyDaySummary, 7)
	dayIndexByKey := make(map[string]int, 7)
	for i := range days {
		current := weekStart.AddDate(0, 0, i)
		days[i] = WeeklyDaySummary{Date: current.Format("2006-01-02")}
		dayIndexByKey[dateKey(current)] = i
	}

	for _, record := range records {
		start := record.StartTime.In(loc)
		index, ok := dayIndexByKey[dateKey(start)]
		if !ok {
			continue
		}

		report.TotalSessions++
		switch record.Status {
		case pomodoro.StatusCompleted:
			report.CompletedSessions++
			report.CompletedFocus += record.PlannedFocus
			days[index].CompletedSessions++
			days[index].CompletedFocus += record.PlannedFocus
		case pomodoro.StatusInterrupted:
			report.InterruptedSessions++
			days[index].InterruptedSessions++
		}
	}

	report.Days = days
	return report
}

func FormatDailySummary(summary DailySummary) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "Daily Summary %s\n", summary.Date)
	fmt.Fprintf(builder, "Total sessions: %d\n", summary.TotalSessions)
	fmt.Fprintf(builder, "Completed sessions: %d\n", summary.CompletedSessions)
	fmt.Fprintf(builder, "Interrupted sessions: %d\n", summary.InterruptedSessions)
	fmt.Fprintf(builder, "Completed focus time: %s\n", summary.CompletedFocus)
	if len(summary.Tasks) == 0 {
		builder.WriteString("Tasks: none\n")
	} else {
		fmt.Fprintf(builder, "Tasks: %s\n", strings.Join(summary.Tasks, ", "))
	}
	return builder.String()
}

func FormatWeeklyReport(report WeeklyReport) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "Weekly Report %s to %s\n", report.WeekStart, report.WeekEnd)
	fmt.Fprintf(builder, "Total sessions: %d\n", report.TotalSessions)
	fmt.Fprintf(builder, "Completed sessions: %d\n", report.CompletedSessions)
	fmt.Fprintf(builder, "Interrupted sessions: %d\n", report.InterruptedSessions)
	fmt.Fprintf(builder, "Completed focus time: %s\n", report.CompletedFocus)
	builder.WriteString("Daily breakdown:\n")
	for _, day := range report.Days {
		fmt.Fprintf(builder, "%s | completed=%d interrupted=%d focus=%s\n", day.Date, day.CompletedSessions, day.InterruptedSessions, day.CompletedFocus)
	}
	return builder.String()
}

func normalizeDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func startOfWeek(value time.Time) time.Time {
	current := normalizeDate(value)
	offset := (int(current.Weekday()) + 6) % 7
	return current.AddDate(0, 0, -offset)
}

func dateKey(value time.Time) string {
	return value.Format("2006-01-02")
}
