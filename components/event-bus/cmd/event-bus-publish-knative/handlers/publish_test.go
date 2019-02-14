package handlers

import (
	"testing"
)

type TestCase struct {
	name     string
	expected string
}

func Test_getChannelName(t *testing.T) {
	// prepare test-cases
	testCases := []TestCase{
		{
			name:     "ordercreated",
			expected: "ordercreated",
		},
		{
			name:     "order_created",
			expected: "order_created",
		},
		{
			name:     "order.created",
			expected: "order-created",
		},
		{
			name:     "order-created",
			expected: "order-created",
		},
	}

	// run test-cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := getChannelName(&testCase.name); got != testCase.expected {
				t.Errorf("getChannelName('%s') got:%s but expected:%s", testCase.name, got, testCase.expected)
			}
		})
	}
}
