# BUG REPRO

## Baseline

- Base branch: `base_bug_004`
- Symptom: when `CompletedFocusSinceLongBreak` reaches `LongBreakEvery` exactly, `RunFocus` still returns a short break

## Reproduction

Run:

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./internal/pomodoro
```

## Expected

`TestRunFocusUsesLongBreakWhenThresholdIsReached` should pass with:

- `NextBreak = long`

## Actual

The test fails because `NextBreak` is `short`
