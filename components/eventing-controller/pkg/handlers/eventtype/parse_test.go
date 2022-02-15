package eventtype

import (
	"testing"
)

func TestParser(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			applicationName, event, version, err := parse(tc.givenEventType, tc.givenPrefix)
			if tc.wantError == true && err == nil {
				t.Fatalf("parse should have failed with an error")
				return
			}
			if tc.wantError != true && err != nil {
				t.Fatalf("parse should have succeeded without an error")
				return
			}
			if tc.wantApplicationName != applicationName {
				t.Fatalf("parse failed event-type[%s] prefix[%s], invalid application name, want [%s] but got [%s]", tc.givenEventType, tc.givenPrefix, tc.wantApplicationName, applicationName)
			}
			if tc.wantEvent != event {
				t.Fatalf("parse failed event-type[%s] prefix[%s], invalid event, want [%s] but got [%s]", tc.givenEventType, tc.givenPrefix, tc.wantEvent, event)
			}
			if tc.wantVersion != version {
				t.Fatalf("parse failed event-type[%s] prefix[%s], invalid version, want [%s] but got [%s]", tc.givenEventType, tc.givenPrefix, tc.wantVersion, version)
			}
		})
	}
}
