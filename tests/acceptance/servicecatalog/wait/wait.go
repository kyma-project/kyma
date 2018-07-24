package wait

import (
	"testing"
	"time"
)

func AtMost(t *testing.T, fn func() error, timeout time.Duration) {
	tick := time.Tick(time.Second)

waitingLoop:
	for {
		select {
		case <-time.After(timeout):
			t.Fatalf("Waiting for resource failed in given timeout %v", timeout)
		case <-tick:
			if err := fn(); err != nil {
				t.Log(err)
				continue
			}
			break waitingLoop
		}
	}
}
