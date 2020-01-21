package waiter

import (
	"fmt"
	"time"
)

func WaitAtMost(fn func() (bool, error), duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(500 * time.Millisecond)

	for {
		ok, err := fn()
		select {
		case <-timeout:
			return fmt.Errorf("waiting for resource failed in given timeout %f second(s)", duration.Seconds())
		case <-tick:
			if err != nil {
				return err
			} else if ok {
				return nil
			}
		}
	}
}
