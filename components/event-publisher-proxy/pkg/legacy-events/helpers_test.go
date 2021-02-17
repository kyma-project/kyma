package legacy

import (
	"testing"
)

func TestParseApplicationNameFromPath(t *testing.T) {
	testCases := []struct {
		name          string
		inputPath     string
		wantedAppName string
	}{
		{
			name:          "should return application when correct path is used",
			inputPath:     "/application/v1/events",
			wantedAppName: "application",
		}, {
			name:          "should return application when extra slash is in the path",
			inputPath:     "//application/v1/events",
			wantedAppName: "application",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotAppName := ParseApplicationNameFromPath(tc.inputPath)
			if tc.wantedAppName != gotAppName {
				t.Errorf("incorrect parsing, wanted: %s, got: %s", tc.wantedAppName, gotAppName)
			}
		})
	}
}

func TestFormatEventType4BEB(t *testing.T) {
	testCases := []struct {
		eventTypePrefix string
		app             string
		eventType       string
		version         string
		wantedEventType string
	}{
		{
			eventTypePrefix: "prefix",
			app:             "app",
			eventType:       "order.foo",
			version:         "v1",
			wantedEventType: "prefix.app.order.foo.v1",
		},
		{
			eventTypePrefix: "prefix",
			app:             "app",
			eventType:       "order-foo",
			version:         "v1",
			wantedEventType: "prefix.app.order-foo.v1",
		},
	}

	for _, tc := range testCases {
		gotEventType := formatEventType4BEB(tc.eventTypePrefix, tc.app, tc.eventType, tc.version)
		if tc.wantedEventType != gotEventType {
			t.Errorf("incorrect formatting of eventType: "+
				"%s, wanted: %s got: %s", tc.eventType, tc.wantedEventType, gotEventType)
		}
	}
}
