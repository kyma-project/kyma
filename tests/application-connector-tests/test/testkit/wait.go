package testkit

import (
	"errors"
	"time"
)

func WaitForFunction(interval, timeout time.Duration, conditionalFunc func() bool) error {
	done := time.After(timeout)

	for {
		if conditionalFunc() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for condition")
		default:
			time.Sleep(interval)
		}
	}
}
