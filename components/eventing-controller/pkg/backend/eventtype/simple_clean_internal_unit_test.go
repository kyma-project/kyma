package eventtype

import (
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

func TestSimpleCleaner(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenEventType       string
		wantEventType        string
		wantError            bool
	}{
		{
			name:                 "success if prefix is empty",
			givenEventTypePrefix: "",
			givenEventType:       "testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the application name is clean",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.testapp.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the application name needs to be cleaned",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.te--s__t!!a@@p##p%%.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the application name needs to be cleaned and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenEventType: "prefix.te--s__t!!a@@p##p%%.Segment1.Segment2.Segment3." +
				"Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType: "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:     false,
		},
		{
			name:                 "success if the given application name is clean and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.test-app.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is not clean",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is not clean and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.testapp.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		// invalid even-types
		{
			name:                 "fail if the prefix is invalid",
			givenEventTypePrefix: "prefix",
			givenEventType:       "invalid.prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantError:            true,
		},
		{
			name:                 "fail if the prefix is missing",
			givenEventTypePrefix: "prefix",
			givenEventType:       "testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantError:            true,
		},
		{
			name:                 "fail if the event-type is incomplete",
			givenEventTypePrefix: "prefix",
			givenEventType:       "prefix.testapp.Segment1-Part1-Part2-Ä.v1",
			wantError:            true,
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

			cleaner := NewSimpleCleaner(tc.givenEventTypePrefix, defaultLogger)
			eventType, err := cleaner.Clean(tc.givenEventType)
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantEventType, eventType)
			}
		})
	}
}
