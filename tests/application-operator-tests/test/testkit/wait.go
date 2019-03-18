package testkit

import (
	"errors"
	"time"
)

func (ts *TestSuite) waitForFunction(interval, timeout time.Duration, conditionalFunc func() bool) error {
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

func (ts *TestSuite) shouldLastFor(interval, timeout time.Duration, conditionalFunc func() bool) error {
	done := time.After(timeout)

	for {
		if !conditionalFunc() {
			return errors.New("unexpected condition occurred when should not")
		}

		select {
		case <-done:
			if !conditionalFunc() {
				return errors.New("unexpected condition occurred when should not")
			}

			return nil
		default:
			time.Sleep(defaultCheckInterval)
		}
	}
}
