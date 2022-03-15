package eventtype

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
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
		// valid even-types for existing applications
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
			name:                   "success if the given application type label is clean",
			givenEventTypePrefix:   "prefix",
			givenApplicationName:   "testapp",
			givenApplicationLabels: map[string]string{application.TypeLabel: "testapptype"},
			givenEventType:         "prefix.testapp.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:          "prefix.testapptype.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:              false,
		},
		{
			name:                   "success if the given application type label is clean and event has more than two segments",
			givenEventTypePrefix:   "prefix",
			givenApplicationName:   "testapp",
			givenApplicationLabels: map[string]string{application.TypeLabel: "testapptype"},
			givenEventType:         "prefix.testapp.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:          "prefix.testapptype.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:              false,
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
		{
			name:                   "success if the given application type label needs to be cleaned",
			givenEventTypePrefix:   "prefix",
			givenApplicationName:   "te--s__t!!a@@p##p%%",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			givenEventType:         "prefix.te--s__t!!a@@p##p%%.Segment1-Part1-Part2-Ä.Segment2-Part1-Part2-Ä.v1",
			wantEventType:          "prefix.testapptype.Segment1Part1Part2.Segment2Part1Part2.v1",
			wantError:              false,
		},
		{
			name:                   "success if the given application type label needs to be cleaned and event has more than two segments",
			givenEventTypePrefix:   "prefix",
			givenApplicationName:   "te--s__t!!a@@p##p%%",
			givenApplicationLabels: map[string]string{application.TypeLabel: "t..e--s__t!!a@@p##p%%t^^y&&p**e"},
			givenEventType:         "prefix.te--s__t!!a@@p##p%%.Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
			wantEventType:          "prefix.testapptype.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
			wantError:              false,
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

	logger := logrus.New()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := applicationtest.NewApplication(tc.givenApplicationName, tc.givenApplicationLabels)
			appLister := fake.NewApplicationListerOrDie(context.Background(), app)
			cleaner := NewCleaner(tc.givenEventTypePrefix, appLister, logger)
			eventType, err := cleaner.Clean(tc.givenEventType)
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantEventType, eventType)
			}
		})
	}
}
