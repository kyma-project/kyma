package testkit

import (
	"fmt"
	"time"
)

func WaitUntil(retries int, sleepTimeSeconds int, predicate func() (bool, error)) error {
	var ready bool
	var e error

	sleepDuration := time.Duration(sleepTimeSeconds) * time.Second

	for i := 0; i < retries; i++ {
		ready, e = predicate()
		if e != nil {
			return e
		}
		if ready {
			break
		}
		time.Sleep(sleepDuration)
	}

	if ready {
		return nil
	}

	return fmt.Errorf("resource not ready")
}
