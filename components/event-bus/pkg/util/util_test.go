package util_test

import (
	"strings"
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
			name:             "test-case-1",
			sourceID:         "ec-default",
			eventType:        "order.created",
			eventTypeVersion: "v1",
			//channel name should starts with "k"+ timestamp
			expectedChannelName: "k0000000000-ec-default-ord",
		},
	}

	// run the test-cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			channelName := util.GetKnativeChannelName(&testCase.sourceID, &testCase.eventType, 25)
			tokens := strings.Split(channelName, "-")
			if !strings.HasPrefix(tokens[0], "k") || tokens[1] != "ec" || tokens[2] != "default" ||
				tokens[3] != "ord" || len(channelName) != 26 {
				t.Errorf("GetChannelName generates chanel name [%s] but expected [%s] with the current timestamp",
					channelName, testCase.expectedChannelName)
			}
		})
	}
}
