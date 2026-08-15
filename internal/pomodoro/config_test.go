package pomodoro

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default config to be valid, got %v", err)
	}

	cfg.FocusDuration = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected zero focus duration to fail validation")
	}
}

func TestParseDurationInput(t *testing.T) {
	duration, err := ParseDurationInput("45m")
	if err != nil {
		t.Fatalf("expected duration parse to succeed, got %v", err)
	}
	if duration != 45*time.Minute {
		t.Fatalf("expected 45 minutes, got %s", duration)
	}

	if _, err := ParseDurationInput("0m"); err == nil {
		t.Fatal("expected zero duration to fail")
	}
}
