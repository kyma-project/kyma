package testkit

import (
	"github.com/stretchr/testify/require"
	"strings"
	"time"
)

type conditionFunction struct {
	condition   func() bool
	failMessage string
}

func (ts *TestSuite) waitForFunction(conditionalFunc func() bool, message string, timeout time.Duration) {
	done := time.After(timeout)

	for {
		if conditionalFunc() {
			return
		}

		select {
		case <-done:
			require.Fail(ts.t, message)
		default:
			time.Sleep(defaultCheckInterval)
		}
	}
}

func (ts *TestSuite) waitForFunctions(conditionalFuncs []conditionFunction, timeout time.Duration) {
	done := time.After(timeout)

	for {
		success := true
		failMessages := []string{}

		for _, condFunc := range conditionalFuncs {
			result := condFunc.condition()

			if result == false {
				failMessages = append(failMessages, condFunc.failMessage)
				success = false
			}
		}

		if success {
			return
		}

		select {
		case <-done:
			failMessage := strings.Join(failMessages, "\n")
			require.Fail(ts.t, failMessage)
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
