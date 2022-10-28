package eventtype

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                string
		givenEventType      string
		givenPrefix         string
		wantApplicationName string
		wantEvent           string
		wantVersion         string
		wantError           bool
	}{
		{
			name:           "should fail if prefix is missing",
			givenEventType: "prefix.test.app.name.123.order.created.v1",
			givenPrefix:    "missing.prefix",
			wantError:      true,
		},
		{
			name:           "should fail if prefix is duplicated",
			givenEventType: "one.two.one.two.prefix.test.app.name.123.order.created.v1",
			givenPrefix:    "one.two",
			wantError:      true,
		},
		{
			name:           "should fail if event-type is incomplete",
			givenEventType: "prefix.order.created.v1",
			givenPrefix:    "prefix",
			wantError:      true,
		},
		{
			name:                "should succeed if prefix is empty",
			givenEventType:      "application.order.created.v1",
			givenPrefix:         "",
			wantApplicationName: "application",
			wantEvent:           "order.created",
			wantVersion:         "v1",
			wantError:           false,
		},
		{
			name:                "should succeed if event has two segments",
			givenEventType:      "prefix.test_app-123.Segment1.Segment2.v1",
			givenPrefix:         "prefix",
			wantApplicationName: "test_app-123",
			wantEvent:           "Segment1.Segment2",
			wantVersion:         "v1",
			wantError:           false,
		},
		{
			name:                "should succeed if event has more than two segments",
			givenEventType:      "prefix.test_app-123.Segment1.Segment2.Segment3.Segment4.Segment5.v1",
			givenPrefix:         "prefix",
			wantApplicationName: "test_app-123",
			wantEvent:           "Segment1Segment2Segment3Segment4.Segment5",
			wantVersion:         "v1",
			wantError:           false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			applicationName, event, version, err := parse(tc.givenEventType, tc.givenPrefix)
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantApplicationName, applicationName)
				require.Equal(t, tc.wantEvent, event)
				require.Equal(t, tc.wantVersion, version)
			}
		})
	}
}

func Test_checkForEmptySegments(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		givenSegments []string
		wantResult    bool
	}{
		{
			name:          "should pass if all segments are non-empty",
			givenSegments: []string{"one", "two", "three"},
			wantResult:    false,
		},
		{
			name:          "should fail if any segment is empty",
			givenSegments: []string{"one", "", "three"},
			wantResult:    true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantResult, checkForEmptySegments(tc.givenSegments))
		})
	}
}
