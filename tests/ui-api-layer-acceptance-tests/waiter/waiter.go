package waiter

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

func WaitAtMost(fn func() (bool, error), duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return errors.New(fmt.Sprintf("Waiting for resource failed in given timeout %f second(s)", duration.Seconds()))
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
