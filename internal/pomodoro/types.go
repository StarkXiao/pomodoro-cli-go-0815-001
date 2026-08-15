package pomodoro

import "time"

const (
	StatusCompleted   = "completed"
	StatusInterrupted = "interrupted"

	BreakShort = "short"
	BreakLong  = "long"
)

type Config struct {
	FocusDuration      time.Duration `json:"-"`
	ShortBreakDuration time.Duration `json:"-"`
	LongBreakDuration  time.Duration `json:"-"`
	LongBreakEvery     int           `json:"long_break_every"`
}

type configJSON struct {
	FocusDuration      string `json:"focus_duration"`
	ShortBreakDuration string `json:"short_break_duration"`
	LongBreakDuration  string `json:"long_break_duration"`
	LongBreakEvery     int    `json:"long_break_every"`
}

type RuntimeState struct {
	CompletedFocusSinceLongBreak int       `json:"completed_focus_since_long_break"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

type SessionRecord struct {
	ID           string        `json:"id"`
	TaskName     string        `json:"task_name"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	PlannedFocus time.Duration `json:"-"`
	Status       string        `json:"status"`
	NextBreak    string        `json:"next_break,omitempty"`
}

type sessionJSON struct {
	ID           string    `json:"id"`
	TaskName     string    `json:"task_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	PlannedFocus string    `json:"planned_focus"`
	Status       string    `json:"status"`
	NextBreak    string    `json:"next_break,omitempty"`
}

type DataFile struct {
	Config   Config          `json:"-"`
	Runtime  RuntimeState    `json:"runtime"`
	Sessions []SessionRecord `json:"-"`
}

type dataFileJSON struct {
	Config   configJSON    `json:"config"`
	Runtime  RuntimeState  `json:"runtime"`
	Sessions []sessionJSON `json:"sessions"`
}

func DefaultConfig() Config {
	return Config{
		FocusDuration:      25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakEvery:     4,
	}
}
