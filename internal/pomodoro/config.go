package pomodoro

import (
	"fmt"
	"strings"
	"time"
)

func (c Config) Validate() error {
	if c.FocusDuration <= 0 {
		return fmt.Errorf("focus duration must be greater than zero")
	}
	if c.ShortBreakDuration <= 0 {
		return fmt.Errorf("short break duration must be greater than zero")
	}
	if c.LongBreakDuration <= 0 {
		return fmt.Errorf("long break duration must be greater than zero")
	}
	if c.LongBreakEvery <= 0 {
		return fmt.Errorf("long break rule must trigger every at least 1 focus session")
	}
	return nil
}

func ParseDurationInput(value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	duration, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", value, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be greater than zero")
	}
	return duration, nil
}

func (c Config) toJSON() configJSON {
	return configJSON{
		FocusDuration:      c.FocusDuration.String(),
		ShortBreakDuration: c.ShortBreakDuration.String(),
		LongBreakDuration:  c.LongBreakDuration.String(),
		LongBreakEvery:     c.LongBreakEvery,
	}
}

func configFromJSON(in configJSON) (Config, error) {
	focus, err := ParseDurationInput(in.FocusDuration)
	if err != nil {
		return Config{}, fmt.Errorf("parse focus duration: %w", err)
	}
	shortBreak, err := ParseDurationInput(in.ShortBreakDuration)
	if err != nil {
		return Config{}, fmt.Errorf("parse short break duration: %w", err)
	}
	longBreak, err := ParseDurationInput(in.LongBreakDuration)
	if err != nil {
		return Config{}, fmt.Errorf("parse long break duration: %w", err)
	}

	cfg := Config{
		FocusDuration:      focus,
		ShortBreakDuration: shortBreak,
		LongBreakDuration:  longBreak,
		LongBreakEvery:     in.LongBreakEvery,
	}
	return cfg, cfg.Validate()
}
