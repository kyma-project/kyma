package util

import (
	"testing"
)

func Test_getChannelName(t *testing.T) {
	// define the test-case struct
	type TestCase struct {
		name             string
		sourceID         string
		eventType        string
		eventTypeVersion string
		expected         string
	}

	// cache the test-cases output to detect naming collisions
	channelNames := make(map[string]TestCase)

	// initialize the test-cases
	testCases := []TestCase{
		{
			name:             "test-case-1",
			sourceID:         "ec-default",
			eventType:        "order.created",
			eventTypeVersion: "v1",
			expected:         "ec--default-order-dot-created-v1",
		},
		{
			name:             "test-case-2",
			sourceID:         "ec-default",
			eventType:        "order-created",
			eventTypeVersion: "v1",
			expected:         "ec--default-order--created-v1",
		},
		{
			name:             "test-case-3",
			sourceID:         "ec.default",
			eventType:        "order.created",
			eventTypeVersion: "v1",
			expected:         "ec-dot-default-order-dot-created-v1",
		},
		{
			name:             "test-case-4",
			sourceID:         "ec.default",
			eventType:        "order-created",
			eventTypeVersion: "v1",
			expected:         "ec-dot-default-order--created-v1",
		},
		{
			name:             "test-case-5",
			sourceID:         "ec..default--test..1",
			eventType:        "order--created..test--1",
			eventTypeVersion: "v1",
			expected:         "ec-dot--dot-default----test-dot--dot-1-order----created-dot--dot-test----1-v1",
		},
		{
			name:             "test-case-6",
			sourceID:         "ec--default..test--1",
			eventType:        "order..created--test..1",
			eventTypeVersion: "v1",
			expected:         "ec----default-dot--dot-test----1-order-dot--dot-created----test-dot--dot-1-v1",
		},
		{
			name:             "test-case-7",
			sourceID:         "external-application",
			eventType:        "order-created",
			eventTypeVersion: "v1",
			expected:         "external--application-order--created-v1",
		},
		{
			name:             "test-case-8",
			sourceID:         "external-application-order",
			eventType:        "created",
			eventTypeVersion: "v1",
			expected:         "external--application--order-created-v1",
		},
		{
			name:             "test-case-9",
			sourceID:         "external.application",
			eventType:        "order.created",
			eventTypeVersion: "v1",
			expected:         "external-dot-application-order-dot-created-v1",
		},
		{
			name:             "test-case-10",
			sourceID:         "external.application.order",
			eventType:        "created",
			eventTypeVersion: "v1",
			expected:         "external-dot-application-dot-order-created-v1",
		},
	}

	// run the test-cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := GetChannelName(&testCase.sourceID, &testCase.eventType, &testCase.eventTypeVersion)

			// check the channel naming correctness
			if result != testCase.expected {
				t.Errorf("getChannelName returned [%s] but expected [%s]", result, testCase.expected)
			}

			// check the channel naming collisions
			if tc, found := channelNames[result]; found {
				t.Errorf("getChannelName generates the same chanel name [%s] for test case [%s] and [%s]", result, tc.name, testCase.name)
			}

			// cache the test result
			channelNames[result] = testCase
		})
	}
}
