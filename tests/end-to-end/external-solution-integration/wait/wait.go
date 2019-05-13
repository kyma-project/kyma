package wait

import (
	"errors"
	"time"
)

func Until(retries int, sleepTimeSeconds int, predicate func() (bool, error)) error {
	var ready bool
	var e error

	sleepDuration := time.Duration(sleepTimeSeconds) * time.Second

	for i := 0; i < retries && !ready; i++ {
		ready, e = predicate()
		if e != nil {
			return e
		}
		time.Sleep(sleepDuration)
	}

	if ready {
		return nil
	}

	return errors.New("resource not ready")
}
