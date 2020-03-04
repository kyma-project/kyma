package testkit

import (
	"errors"
	"time"
)

func WaitForFunction(interval, timeout time.Duration, isReady func() bool) error {
	done := time.After(timeout)

	for {
		if isReady() {
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

func ShouldLastFor(interval, timeout time.Duration, conditionalFunc func() bool) error {
	done := time.After(timeout)

	for {
		if !conditionalFunc() {
			return errors.New("unexpected condition occurred")
		}

		select {
		case <-done:
			if !conditionalFunc() {
				return errors.New("unexpected condition occurred")
			}

			return nil
		default:
			time.Sleep(interval)
		}
	}
}
