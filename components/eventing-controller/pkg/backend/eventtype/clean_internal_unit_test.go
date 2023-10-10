package eventtype

import (
	"testing"

	"github.com/stretchr/testify/require"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

func TestCleaner(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                   string
		givenEventTypePrefix   string
		givenApplicationName   string
		givenApplicationLabels map[string]string
		givenEventType         string
		wantEventType          string
		wantError              bool
	}{
		// valid even-types
		{
			name:                 "success if prefix is empty",
			givenEventTypePrefix: "",
			givenApplicationName: "testapp",
			givenEventType:       "testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is clean",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "testapp",
			givenEventType:       "prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is clean and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "testapp",
			givenEventType:       "prefix.testapp.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name needs to be cleaned",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "te--s__t!!a@@p##p%%",
			givenEventType:       "prefix.te--s__t!!a@@p##p%%.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name needs to be cleaned and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "te--s__t!!a@@p##p%%",
			givenEventType:       "prefix.te--s__t!!a@@p##p%%.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		// valid even-types for non-existing applications to simulate in-cluster eventing
		{
			name:                 "success if the given application name is clean for non-existing application",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "",
			givenEventType:       "prefix.test-app.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is clean for non-existing application and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "",
			givenEventType:       "prefix.test-app.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is not clean for non-existing application",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "",
			givenEventType:       "prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:            false,
		},
		{
			name:                 "success if the given application name is not clean for non-existing application and event has more than two segments",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "",
			givenEventType:       "prefix.testapp.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:        "prefix.testapp.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:            false,
		},
		// invalid even-types
		{
			name:                 "fail if the prefix is invalid",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "testapp",
			givenEventType:       "invalid.prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantError:            true,
		},
		{
			name:                 "fail if the prefix is missing",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "testapp",
			givenEventType:       "testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantError:            true,
		},
		{
			name:                 "fail if the event-type is incomplete",
			givenEventTypePrefix: "prefix",
			givenApplicationName: "testapp",
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
			cleaner := NewCleaner(tc.givenEventTypePrefix, defaultLogger)
			eventType, err := cleaner.Clean(tc.givenEventType)
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantEventType, eventType)
			}
		})
	}
}
