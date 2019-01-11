package testkit

import (
	"time"
)

func (ts *TestSuite) waitForFunction(conditionalFunc func() bool, message string, timeout time.Duration) {
	done := time.After(timeout)

	for {
		if conditionalFunc() {
			return
		}

		select {
		case <-done:
			ts.t.Errorf(message)
		default:
			time.Sleep(defaultCheckInterval)
		}
	}
}

func (ts *TestSuite) shouldLastFor(conditionalFunc func() bool, message string, timeout time.Duration) {
	done := time.After(timeout)

	for {
		if !conditionalFunc() {
			ts.t.Errorf(message)
		}

		select {
		case <-done:
			if !conditionalFunc() {
				ts.t.Errorf(message)
			}
			return
		default:
			time.Sleep(defaultCheckInterval)
		}
	}
}
