package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"localpomodoro/internal/pomodoro"
)

type JSONStore struct {
	Path string
}

type diskConfig struct {
	FocusDuration      string `json:"focus_duration"`
	ShortBreakDuration string `json:"short_break_duration"`
	LongBreakDuration  string `json:"long_break_duration"`
	LongBreakEvery     int    `json:"long_break_every"`
}

type diskSession struct {
	ID           string    `json:"id"`
	TaskName     string    `json:"task_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	PlannedFocus string    `json:"planned_focus"`
	Status       string    `json:"status"`
	NextBreak    string    `json:"next_break,omitempty"`
}

type diskData struct {
	Config   diskConfig            `json:"config"`
	Runtime  pomodoro.RuntimeState `json:"runtime"`
	Sessions []diskSession         `json:"sessions"`
}

func (s JSONStore) Load() (pomodoro.DataFile, error) {
	if s.Path == "" {
		return pomodoro.DataFile{}, fmt.Errorf("store path is required")
	}

	content, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			data := pomodoro.DataFile{
				Config:  pomodoro.DefaultConfig(),
				Runtime: pomodoro.RuntimeState{UpdatedAt: time.Now()},
			}
			return data, data.Config.Validate()
		}
		return pomodoro.DataFile{}, fmt.Errorf("read data file: %w", err)
	}

	var raw diskData
	if err := json.Unmarshal(content, &raw); err != nil {
		return pomodoro.DataFile{}, fmt.Errorf("decode data file: %w", err)
	}

	cfg, err := decodeConfig(raw.Config)
	if err != nil {
		return pomodoro.DataFile{}, err
	}

	data := pomodoro.DataFile{
		Config:  cfg,
		Runtime: raw.Runtime,
	}
	for _, item := range raw.Sessions {
		record, err := decodeSession(item)
		if err != nil {
			return pomodoro.DataFile{}, err
		}
		data.Sessions = append(data.Sessions, record)
	}

	return data, nil
}

func (s JSONStore) Save(data pomodoro.DataFile) error {
	if s.Path == "" {
		return fmt.Errorf("store path is required")
	}
	if err := data.Config.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	payload := diskData{
		Config: diskConfig{
			FocusDuration:      data.Config.FocusDuration.String(),
			ShortBreakDuration: data.Config.ShortBreakDuration.String(),
			LongBreakDuration:  data.Config.LongBreakDuration.String(),
			LongBreakEvery:     data.Config.LongBreakEvery,
		},
		Runtime: data.Runtime,
	}

	for _, session := range data.Sessions {
		if err := session.Validate(); err != nil {
			return fmt.Errorf("validate session %q: %w", session.ID, err)
		}
		payload.Sessions = append(payload.Sessions, diskSession{
			ID:           session.ID,
			TaskName:     session.TaskName,
			StartTime:    session.StartTime,
			EndTime:      session.EndTime,
			PlannedFocus: session.PlannedFocus.String(),
			Status:       session.Status,
			NextBreak:    session.NextBreak,
		})
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode data file: %w", err)
	}
	body = append(body, '\n')

	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(s.Path), "pomodoro-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, s.Path); err != nil {
		return fmt.Errorf("replace data file: %w", err)
	}
	return nil
}

func decodeConfig(in diskConfig) (pomodoro.Config, error) {
	focus, err := pomodoro.ParseDurationInput(in.FocusDuration)
	if err != nil {
		return pomodoro.Config{}, fmt.Errorf("decode config focus duration: %w", err)
	}
	shortBreak, err := pomodoro.ParseDurationInput(in.ShortBreakDuration)
	if err != nil {
		return pomodoro.Config{}, fmt.Errorf("decode config short break duration: %w", err)
	}
	longBreak, err := pomodoro.ParseDurationInput(in.LongBreakDuration)
	if err != nil {
		return pomodoro.Config{}, fmt.Errorf("decode config long break duration: %w", err)
	}

	cfg := pomodoro.Config{
		FocusDuration:      focus,
		ShortBreakDuration: shortBreak,
		LongBreakDuration:  longBreak,
		LongBreakEvery:     in.LongBreakEvery,
	}
	if err := cfg.Validate(); err != nil {
		return pomodoro.Config{}, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func decodeSession(in diskSession) (pomodoro.SessionRecord, error) {
	duration, err := pomodoro.ParseDurationInput(in.PlannedFocus)
	if err != nil {
		return pomodoro.SessionRecord{}, fmt.Errorf("decode session planned focus: %w", err)
	}

	record := pomodoro.SessionRecord{
		ID:           in.ID,
		TaskName:     in.TaskName,
		StartTime:    in.StartTime,
		EndTime:      in.EndTime,
		PlannedFocus: duration,
		Status:       in.Status,
		NextBreak:    in.NextBreak,
	}
	if err := record.Validate(); err != nil {
		return pomodoro.SessionRecord{}, fmt.Errorf("validate session %q: %w", record.ID, err)
	}
	return record, nil
}
