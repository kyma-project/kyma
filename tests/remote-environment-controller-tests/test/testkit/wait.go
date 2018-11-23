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

func (ts *TestSuite) waitForFunction(conditionalFunc func() bool, message string) {
	done := time.After(ts.installationTimeout)

	for {
		if conditionalFunc() {
			return
		}

		select {
		case <-done:
			require.Fail(ts.t, message)
		default:
			time.Sleep(ts.installationCheckInterval)
		}
	}
}

func (ts *TestSuite) waitForFunctions(conditionalFuncs []conditionFunction) {
	done := time.After(ts.installationTimeout)

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
			time.Sleep(ts.installationCheckInterval)
		}
	}
}
