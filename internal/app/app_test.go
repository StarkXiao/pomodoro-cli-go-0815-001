package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConfigSetAndShowWithDataOverride(t *testing.T) {
	dataPath := filepath.Join(t.TempDir(), "pomodoro.json")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{
		"config", "set",
		"--focus", "1m",
		"--short-break", "2m",
		"--long-break", "3m",
		"--long-break-every", "2",
		"--data", dataPath,
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected config set to succeed, got %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"config", "show", "--data", dataPath}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected config show to succeed, got %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"Focus duration: 1m0s",
		"Short break duration: 2m0s",
		"Long break duration: 3m0s",
		"Long break every: 2",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestRunUnlockRemovesLockFile(t *testing.T) {
	workingDir := t.TempDir()
	lockFile := filepath.Join(workingDir, "data", "pomodoro.json.lock")
	if err := os.MkdirAll(filepath.Dir(lockFile), 0o755); err != nil {
		t.Fatalf("expected lock directory creation to succeed, got %v", err)
	}
	if err := os.WriteFile(lockFile, []byte("123\n"), 0o644); err != nil {
		t.Fatalf("expected lock file creation to succeed, got %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected getwd to succeed, got %v", err)
	}
	defer os.Chdir(originalWD)
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("expected chdir to succeed, got %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = Run([]string{"unlock"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected unlock to succeed, got %v", err)
	}

	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Fatalf("expected lock file to be removed, stat err=%v", err)
	}
	if !strings.Contains(stdout.String(), "Removed lock file") {
		t.Fatalf("expected unlock output, got %q", stdout.String())
	}
}
