package repeat

import (
	"testing"
	"time"
)

func FuncAtMost(t *testing.T, fn func() error, duration time.Duration) {
	t.Helper()

	var (
		tick    = time.Tick(time.Second)
		timeout = time.After(duration)

		lastErr error
		execCnt uint
	)

waitingLoop:
	for {
		select {
		case <-timeout:
			t.Fatalf("Waiting for function failed in given timeout %v. Function was executed %d times. Last error: %v", duration, execCnt, lastErr)
		case <-tick:
			execCnt++

			if lastErr = fn(); lastErr != nil {
				continue
			}

			break waitingLoop
		}
	}
}
