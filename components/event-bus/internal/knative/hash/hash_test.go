package hash

import (
	"strings"
	"testing"
)

func Test_hash(t *testing.T) {
	const maxHashLength = 32

	// prepare testCases
	testCases := map[string]string{
		"testCase-0": "source-id--event-type--event-type-version",
		"testCase-1": "source-did--event-dtype--event-dtype-dversion",
	}

	// run tests
	for k, v := range testCases {
		t.Run(k, func(t *testing.T) {
			h := ComputeHash(&v)

			// check hash length
			length := len(h)
			if length != maxHashLength {
				t.Errorf("%s: invalid hash length, expected: [%d] but found: [%d]", k, maxHashLength, length)
			}

			// check hash to be lower case only
			hLower := strings.ToLower(h)
			if h != hLower {
				t.Errorf("%s: invalid character case, expected: [%s] but found: [%s]", k, hLower, h)
			}
		})
	}
}
