package waiter

import (
	"log"
	"time"

	"github.com/pkg/errors"
)

func WaitAtMost(fn func() (bool, error), duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(500 * time.Millisecond)

	for {
		ok, err := fn()
		select {
		case <-timeout:
			return errors.Wrapf(err, "Waiting for resource failed in given timeout %f second(s)", duration.Seconds())
		case <-tick:
			if err != nil {
				log.Println(err)
			} else if ok {
				return nil
			}
		}
	}
}
