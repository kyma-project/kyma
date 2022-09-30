package cleaner

import (
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/stretchr/testify/require"
)

func Test_JSCleanSource(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		givenEventSource string
		wantEventSource  string
	}{
		{
			name:             "removes only dots, greater than, asterisk, spaces and tabs",
			givenEventSource: "source1-Part1-Part*2-Ä.te--<>  <>s__t!!a@  \t @*p##p%%",
			wantEventSource:  "source1-Part1-Part2-Äte--<<s__t!!a@@p##p%%",
		},
		{
			name:             "does nothing for allowed characters",
			givenEventSource: "t!!a@@p##p%Ä",
			wantEventSource:  "t!!a@@p##p%Ä",
		},
	}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cleaner := NewJetStreamCleaner(defaultLogger)
			eventType, _ := cleaner.CleanSource(tc.givenEventSource)
			require.Equal(t, tc.wantEventSource, eventType)
		})
	}
}

func Test_JSCleanEventType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
	}{
		{
			name:           "removes only asterisk, greater than, spaces and tabs",
			givenEventType: "prefix.  testapp.\tSegment1.Segment2>*.Segment3<.Segment4>>-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:  "prefix.testapp.Segment1.Segment2.Segment3<.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
		},
		{
			name:           "does nothing for allowed characters",
			givenEventType: "order!!.created<<.v1##",
			wantEventType:  "order!!.created<<.v1##",
		},
	}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cleaner := NewJetStreamCleaner(defaultLogger)
			eventType, _ := cleaner.CleanEventType(tc.givenEventType)
			require.Equal(t, tc.wantEventType, eventType)
		})
	}
}
