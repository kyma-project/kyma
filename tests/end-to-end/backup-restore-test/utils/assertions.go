package utils

import (
	"fmt"
	"strings"
	"time"
)

// ShouldReturnSubstringEventually receives a function returning a string and compares its output with the expected value
// the function is called at an interval until the timeout kicks in.
func ShouldReturnSubstringEventually(actual interface{}, expected ...interface{}) string {
	if fail := needBetween(3, 4, expected); fail != "" {
		return fail
	}

	action, _ := actual.(func() string)

	if action == nil {
		return "the function needs to return a string \"func() string\""
	}

	until, _ := expected[1].(time.Duration)
	interval, _ := expected[2].(time.Duration)
	substring, _ := expected[0].(string)

	timeout := time.After(until)
	tick := time.Tick(interval)
	debugMessage := ""
	for {
		select {
		case <-timeout:
			//debugMessage += "\n" + printLogsFunctionPodContainers(namespace, functionName)
			if len(expected) == 4 {
				timeoutFunction, ok := expected[3].(func() string)
				if ok {
					debugMessage += timeoutFunction()
				}
			}
			return fmt.Sprintf("Timeout: %v", debugMessage)

		case <-tick:
			value := action()
			debugMessage += "\n" + value
			if strings.Contains(value, substring) {
				return ""
			}
		}
	}
}
