package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"localpomodoro/internal/pomodoro"
	"localpomodoro/internal/report"
	"localpomodoro/internal/storage"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	dataPath, err := resolveDataPath()
	if err != nil {
		return err
	}

	switch args[0] {
	case "start":
		return runStart(args[1:], stdout, dataPath)
	case "config":
		return runConfig(args[1:], stdout, dataPath)
	case "summary":
		return runSummary(args[1:], stdout, dataPath)
	case "report":
		return runReport(args[1:], stdout, dataPath)
	case "unlock":
		return runUnlock(args[1:], stdout, dataPath)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runStart(args []string, stdout io.Writer, dataPath string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	taskName := fs.String("task", "", "task name")
	focusInput := fs.String("focus", "", "focus duration, for example 25m or 45m")
	shortBreakInput := fs.String("short-break", "", "short break duration")
	longBreakInput := fs.String("long-break", "", "long break duration")
	longBreakEveryInput := fs.String("long-break-every", "", "trigger long break after N completed focus sessions")
	dataOverride := fs.String("data", dataPath, "path to JSON data file")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}
	if strings.TrimSpace(*taskName) == "" {
		return fmt.Errorf("task name is required")
	}

	store := storage.JSONStore{Path: *dataOverride}
	lock, err := storage.AcquireFileLock(*dataOverride + ".lock")
	if err != nil {
		if errors.Is(err, storage.ErrDataLocked) {
			return fmt.Errorf("another pomodoro command is already using %s", *dataOverride)
		}
		return err
	}
	defer lock.Release()

	data, err := store.Load()
	if err != nil {
		return err
	}

	cfg, err := applyConfigOverrides(data.Config, *focusInput, *shortBreakInput, *longBreakInput, *longBreakEveryInput)
	if err != nil {
		return err
	}

	service := pomodoro.NewService()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(stdout, "Starting focus timer for %q (%s)\n", strings.TrimSpace(*taskName), cfg.FocusDuration)
	record, runtime, err := service.RunFocus(ctx, stdout, data.Runtime, cfg, *taskName)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			record.EndTime = time.Now()
			record.Status = pomodoro.StatusInterrupted
			data.Sessions = append(data.Sessions, record)
			data.Runtime = runtime
			if saveErr := store.Save(data); saveErr != nil {
				return fmt.Errorf("focus interrupted: %w; failed to save interrupted session: %v", err, saveErr)
			}
			return fmt.Errorf("focus interrupted and saved")
		}
		return err
	}

	data.Sessions = append(data.Sessions, record)
	data.Runtime = runtime
	if err := store.Save(data); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Completed task %q at %s\n", record.TaskName, record.EndTime.Format(time.RFC3339))
	switch record.NextBreak {
	case pomodoro.BreakLong:
		fmt.Fprintf(stdout, "Next break: long break (%s)\n", cfg.LongBreakDuration)
	case pomodoro.BreakShort:
		fmt.Fprintf(stdout, "Next break: short break (%s)\n", cfg.ShortBreakDuration)
	}
	fmt.Fprintf(stdout, "Saved session to %s\n", *dataOverride)
	return nil
}

func runConfig(args []string, stdout io.Writer, dataPath string) error {
	if len(args) == 0 {
		return fmt.Errorf("config subcommand is required: show or set")
	}

	switch args[0] {
	case "show":
		fs := flag.NewFlagSet("config show", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		dataOverride := fs.String("data", dataPath, "path to JSON data file")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() > 0 {
			return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
		}

		store := storage.JSONStore{Path: *dataOverride}
		data, err := store.Load()
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Focus duration: %s\n", data.Config.FocusDuration)
		fmt.Fprintf(stdout, "Short break duration: %s\n", data.Config.ShortBreakDuration)
		fmt.Fprintf(stdout, "Long break duration: %s\n", data.Config.LongBreakDuration)
		fmt.Fprintf(stdout, "Long break every: %d\n", data.Config.LongBreakEvery)
		fmt.Fprintf(stdout, "Data file: %s\n", *dataOverride)
		return nil
	case "set":
		fs := flag.NewFlagSet("config set", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		focusInput := fs.String("focus", "", "focus duration")
		shortBreakInput := fs.String("short-break", "", "short break duration")
		longBreakInput := fs.String("long-break", "", "long break duration")
		longBreakEveryInput := fs.String("long-break-every", "", "trigger long break after N completed focus sessions")
		dataOverride := fs.String("data", dataPath, "path to JSON data file")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() > 0 {
			return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
		}

		lock, err := storage.AcquireFileLock(*dataOverride + ".lock")
		if err != nil {
			if errors.Is(err, storage.ErrDataLocked) {
				return fmt.Errorf("another pomodoro command is already using %s", *dataOverride)
			}
			return err
		}
		defer lock.Release()

		store := storage.JSONStore{Path: *dataOverride}
		data, err := store.Load()
		if err != nil {
			return err
		}

		cfg, err := applyConfigOverrides(data.Config, *focusInput, *shortBreakInput, *longBreakInput, *longBreakEveryInput)
		if err != nil {
			return err
		}
		data.Config = cfg
		data.Runtime.UpdatedAt = time.Now()

		if err := store.Save(data); err != nil {
			return err
		}

		fmt.Fprintf(stdout, "Updated configuration in %s\n", *dataOverride)
		return nil
	default:
		return fmt.Errorf("unknown config subcommand %q", args[0])
	}
}

func runSummary(args []string, stdout io.Writer, dataPath string) error {
	if len(args) == 0 || args[0] != "daily" {
		return fmt.Errorf("supported summary command: daily")
	}

	fs := flag.NewFlagSet("summary daily", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dateInput := fs.String("date", "", "date in YYYY-MM-DD, default today")
	dataOverride := fs.String("data", dataPath, "path to JSON data file")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}

	store := storage.JSONStore{Path: *dataOverride}
	data, err := store.Load()
	if err != nil {
		return err
	}

	day, err := parseDateOrToday(*dateInput)
	if err != nil {
		return err
	}

	summary := report.BuildDailySummary(data.Sessions, day, time.Local)
	_, err = fmt.Fprint(stdout, report.FormatDailySummary(summary))
	return err
}

func runReport(args []string, stdout io.Writer, dataPath string) error {
	if len(args) == 0 || args[0] != "weekly" {
		return fmt.Errorf("supported report command: weekly")
	}

	fs := flag.NewFlagSet("report weekly", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dateInput := fs.String("date", "", "any date inside the target week, format YYYY-MM-DD")
	dataOverride := fs.String("data", dataPath, "path to JSON data file")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}

	store := storage.JSONStore{Path: *dataOverride}
	data, err := store.Load()
	if err != nil {
		return err
	}

	day, err := parseDateOrToday(*dateInput)
	if err != nil {
		return err
	}

	weekly := report.BuildWeeklyReport(data.Sessions, day, time.Local)
	_, err = fmt.Fprint(stdout, report.FormatWeeklyReport(weekly))
	return err
}

func runUnlock(args []string, stdout io.Writer, dataPath string) error {
	fs := flag.NewFlagSet("unlock", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dataOverride := fs.String("data", dataPath, "path to JSON data file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}

	lockPath := *dataOverride + ".lock"
	if err := os.Remove(lockPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(stdout, "No lock file found for %s\n", *dataOverride)
			return nil
		}
		return fmt.Errorf("remove lock file: %w", err)
	}

	fmt.Fprintf(stdout, "Removed lock file %s\n", lockPath)
	return nil
}

func applyConfigOverrides(base pomodoro.Config, focusInput string, shortBreakInput string, longBreakInput string, longBreakEveryInput string) (pomodoro.Config, error) {
	cfg := base
	var err error

	if strings.TrimSpace(focusInput) != "" {
		cfg.FocusDuration, err = pomodoro.ParseDurationInput(focusInput)
		if err != nil {
			return pomodoro.Config{}, fmt.Errorf("invalid focus duration: %w", err)
		}
	}
	if strings.TrimSpace(shortBreakInput) != "" {
		cfg.ShortBreakDuration, err = pomodoro.ParseDurationInput(shortBreakInput)
		if err != nil {
			return pomodoro.Config{}, fmt.Errorf("invalid short break duration: %w", err)
		}
	}
	if strings.TrimSpace(longBreakInput) != "" {
		cfg.LongBreakDuration, err = pomodoro.ParseDurationInput(longBreakInput)
		if err != nil {
			return pomodoro.Config{}, fmt.Errorf("invalid long break duration: %w", err)
		}
	}
	if strings.TrimSpace(longBreakEveryInput) != "" {
		cfg.LongBreakEvery, err = strconv.Atoi(strings.TrimSpace(longBreakEveryInput))
		if err != nil {
			return pomodoro.Config{}, fmt.Errorf("invalid long break rule %q: %w", longBreakEveryInput, err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return pomodoro.Config{}, err
	}
	return cfg, nil
}

func parseDateOrToday(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Now(), nil
	}
	day, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", value)
	}
	return day, nil
}

func resolveDataPath() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	return filepath.Join(workingDir, "data", "pomodoro.json"), nil
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Pomodoro CLI")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  pomodoro start -task \"Task name\" [--focus 25m] [--short-break 5m] [--long-break 15m] [--long-break-every 4] [--data ./data/pomodoro.json]")
	fmt.Fprintln(out, "  pomodoro config show [--data ./data/pomodoro.json]")
	fmt.Fprintln(out, "  pomodoro config set [--focus 25m] [--short-break 5m] [--long-break 15m] [--long-break-every 4] [--data ./data/pomodoro.json]")
	fmt.Fprintln(out, "  pomodoro summary daily [--date 2026-08-15] [--data ./data/pomodoro.json]")
	fmt.Fprintln(out, "  pomodoro report weekly [--date 2026-08-15] [--data ./data/pomodoro.json]")
	fmt.Fprintln(out, "  pomodoro unlock [--data ./data/pomodoro.json]")
}
