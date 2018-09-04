package syncer

import (
	"fmt"
	"log"
	"time"
)

// WaitAtMost waits for the given amount of time until the given function will return true
func WaitAtMost(fn func() (bool, error), duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("waiting for resource failed in given timeout %f second(s)", duration.Seconds())
		case <-tick:
			ok, err := fn()
			if err != nil {
				log.Println(err)
			} else if ok {
				return nil
			}
		}
	}
}
