# BUG REPRO

## Baseline

- Base branch: `base_bug_001`
- Symptom: weekly reports start on Sunday instead of Monday

## Reproduction

Run:

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./internal/report
```

## Expected

For the week containing `2026-08-15`, the weekly report should use:

- `WeekStart = 2026-08-10`
- `WeekEnd = 2026-08-16`

## Actual

The failing test reports:

- `WeekStart = 2026-08-09`

## Failure Signal

`TestBuildWeeklyReportUsesMondayAsWeekStart` fails in `internal/report/report_test.go`.
