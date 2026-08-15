package pomodoro

import (
	"context"
	"fmt"
	"io"
	"time"
)

func RunProgress(ctx context.Context, duration time.Duration, out io.Writer) error {
	if duration <= 0 {
		return fmt.Errorf("duration must be greater than zero")
	}

	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	printProgress(out, duration, duration)
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(out)
			return ctx.Err()
		case tickAt := <-ticker.C:
			elapsed := tickAt.Sub(start)
			remaining := duration - elapsed
			if remaining <= 0 {
				printProgress(out, 0, duration)
				fmt.Fprintln(out)
				return nil
			}
			printProgress(out, remaining, duration)
		}
	}
}

func printProgress(out io.Writer, remaining time.Duration, total time.Duration) {
	percent := 100
	if total > 0 {
		progress := total - remaining
		if progress < 0 {
			progress = 0
		}
		percent = int(float64(progress) / float64(total) * 100)
		if percent > 100 {
			percent = 100
		}
	}

	fmt.Fprintf(out, "\rProgress %3d%% | remaining %s", percent, roundToSecond(remaining))
}

func roundToSecond(value time.Duration) time.Duration {
	if value <= 0 {
		return 0
	}
	return value.Truncate(time.Second)
}
