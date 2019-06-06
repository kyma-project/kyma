package util_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/event-bus/pkg/util"
)

func TestGetChannelName(t *testing.T) {
	// define the test-case struct
	type TestCase struct {
		name                string
		sourceID            string
		eventType           string
		eventTypeVersion    string
		expectedChannelName string
	}

	// initialize the test-cases
	testCases := []TestCase{
		{
			name:                "test-case-1",
			sourceID:            "ec-default",
			eventType:           "order.created",
			eventTypeVersion:    "v1",
			expectedChannelName: "kf5wxkg4bobejchlt6ekbpuwiixddqenw",
		},
	}

	// run the test-cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			channelName := util.GetChannelName(&testCase.sourceID, &testCase.eventType, &testCase.eventTypeVersion)
			if channelName != testCase.expectedChannelName {
				t.Errorf("GetChannelName generates chanel name [%s] but expected [%s]", channelName, testCase.expectedChannelName)
			}
		})
	}
}
