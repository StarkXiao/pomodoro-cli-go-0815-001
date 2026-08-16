# BUG REPRO

## Baseline

- Base branch: `base_bug_003`
- Symptom: daily and weekly reports assign cross-midnight focus sessions to the end date instead of the start date

## Reproduction

Run:

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./internal/report
```

## Expected

`TestBuildReportsUseSessionStartDateForCrossMidnightSessions` should pass with:

- `2026-08-15` counted
- `2026-08-16` not counted

## Actual

The test fails because the session is counted on `2026-08-16`
