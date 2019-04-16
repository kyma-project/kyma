package util

import (
	"testing"
)

func Test_encodeChannelName(t *testing.T) {
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
			expected:         "ec-ddefault--order-pcreated--v1",
		},
		{
			name:             "test-case-2",
			sourceID:         "ec-default",
			eventType:        "order-created",
			eventTypeVersion: "v1",
			expected:         "ec-ddefault--order-dcreated--v1",
		},
		{
			name:             "test-case-3",
			sourceID:         "ec.default",
			eventType:        "order.created",
			eventTypeVersion: "v1",
			expected:         "ec-pdefault--order-pcreated--v1",
		},
		{
			name:             "test-case-4",
			sourceID:         "ec.default",
			eventType:        "order-created",
			eventTypeVersion: "v1",
			expected:         "ec-pdefault--order-dcreated--v1",
		},
		{
			name:             "test-case-5",
			sourceID:         "ec..default--test..1",
			eventType:        "order--created..test--1",
			eventTypeVersion: "v1",
			expected:         "ec-p-pdefault-d-dtest-p-p1--order-d-dcreated-p-ptest-d-d1--v1",
		},
		{
			name:             "test-case-6",
			sourceID:         "ec--default..test--1",
			eventType:        "order..created--test..1",
			eventTypeVersion: "v1",
			expected:         "ec-d-ddefault-p-ptest-d-d1--order-p-pcreated-d-dtest-p-p1--v1",
		},
		{
			name:             "test-case-7",
			sourceID:         "external-application",
			eventType:        "order-created",
			eventTypeVersion: "v1",
			expected:         "external-dapplication--order-dcreated--v1",
		},
		{
			name:             "test-case-8",
			sourceID:         "external-application-order",
			eventType:        "created",
			eventTypeVersion: "v1",
			expected:         "external-dapplication-dorder--created--v1",
		},
		{
			name:             "test-case-9",
			sourceID:         "external.application",
			eventType:        "order.created",
			eventTypeVersion: "v1",
			expected:         "external-papplication--order-pcreated--v1",
		},
		{
			name:             "test-case-10",
			sourceID:         "external.application.order",
			eventType:        "created",
			eventTypeVersion: "v1",
			expected:         "external-papplication-porder--created--v1",
		},
	}

	// run the test-cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := encodeChannelName(&testCase.sourceID, &testCase.eventType, &testCase.eventTypeVersion)

			// check the channel name encoding correctness
			if result != testCase.expected {
				t.Errorf("encodeChannelName returned [%s] but expected [%s]", result, testCase.expected)
			}

			// check the channel name encoding collisions
			if tc, found := channelNames[result]; found {
				t.Errorf("encodeChannelName generates the same chanel name [%s] for test case [%s] and [%s]", result, tc.name, testCase.name)
			}

			// cache the test result
			channelNames[result] = testCase
		})
	}
}
