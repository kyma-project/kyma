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
			ts.logAndFail(message)
		default:
			time.Sleep(defaultCheckInterval)
		}
	}
}

func (ts *TestSuite) shouldLastFor(conditionalFunc func() bool, message string, timeout time.Duration) {
	done := time.After(timeout)

	for {
		if !conditionalFunc() {
			ts.logAndFail(message)
		}

		select {
		case <-done:
			if !conditionalFunc() {
				ts.logAndFail(message)
			}
			return
		default:
			time.Sleep(defaultCheckInterval)
		}
	}
}

func (ts *TestSuite) logAndFail(message string) {
	ts.t.Log(message)
	ts.t.Fail()
}
