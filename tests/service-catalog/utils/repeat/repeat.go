package repeat

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func AssertFuncAtMost(t *testing.T, fn func() error, duration time.Duration) {
	t.Helper()
	require.NoError(t, FuncAtMost(fn, duration))
}

func FuncAtMost(fn func() error, duration time.Duration) error {
	var (
		tick    = time.Tick(time.Second)
		timeout = time.After(duration)

		lastErr error
		execCnt uint
	)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("waiting for function failed in given timeout %v. Function was executed %d times. Last error: %v", duration, execCnt, lastErr)
		case <-tick:
			execCnt++

			if lastErr = fn(); lastErr != nil {
				continue
			}

			return nil
		}
	}
}
