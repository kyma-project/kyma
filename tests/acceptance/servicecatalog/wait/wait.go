package wait

import (
	"testing"
	"time"
)

func ForFuncAtMost(t *testing.T, fn func() error, duration time.Duration) {
	tick := time.Tick(time.Second)
	timeout := time.After(duration)

waitingLoop:
	for {
		select {
		case <-timeout:
			t.Fatalf("Waiting for resource failed in given timeout %v", duration)
		case <-tick:
			if err := fn(); err != nil {
				t.Log(err)
				continue
			}
			break waitingLoop
		}
	}
}
