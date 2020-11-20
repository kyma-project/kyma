package legacy

import (
	"testing"
	"time"

	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
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
			gotAppName := parseApplicationNameFromPath(tc.inputPath)
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
			wantedEventType: "prefix.app.order.foo.v1",
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

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	bebNs := "beb.namespace"
	eventTypePrefix := "event.type.prefix"
	legacyTransformer := NewTransformer(bebNs, eventTypePrefix)
	eventID := "id"
	appName := "foo-app"
	legacyEventType := "foo.bar"
	legacyEventVersion := "v1"
	data := "{\"foo\": \"bar\"}"
	timeNow := time.Now()
	expectedEventType := formatEventType4BEB(eventTypePrefix, appName, legacyEventType, legacyEventVersion)
	timeNowStr := timeNow.Format(time.RFC3339)
	timeNowFormatted, _ := time.Parse(time.RFC3339, timeNowStr)
	publishReqParams := &legacyapi.PublishEventParametersV1{
		PublishrequestV1: legacyapi.PublishRequestV1{
			EventID:          eventID,
			EventType:        legacyEventType,
			EventTime:        timeNowStr,
			EventTypeVersion: legacyEventVersion,
			Data:             data,
		},
	}

	gotEvent, err := legacyTransformer.convertPublishRequestToCloudEvent(appName, publishReqParams)
	if err != nil {
		t.Fatal("failed to convert publish request to CE", err)
	}
	if gotEvent.Context.GetID() != eventID {
		t.Errorf("incorrect id, want: %s, got: %s", eventID, gotEvent.Context.GetDataContentType())
	}

	if gotEvent.Context.GetType() != expectedEventType {
		t.Errorf("incorrect eventType, want: %s, got: %s", expectedEventType, gotEvent.Context.GetType())
	}

	if gotEvent.Context.GetTime() != timeNowFormatted {
		t.Errorf("incorrect eventTime, want: %s, got: %s", timeNowFormatted, gotEvent.Context.GetTime())
	}

	if gotEvent.Context.GetSource() != bebNs {
		t.Errorf("incorrect source want: %s, got: %s", bebNs, gotEvent.Context.GetDataContentType())
	}

	if gotEvent.Context.GetDataContentType() != ContentTypeApplicationJSON {
		t.Errorf("incorrect content-type, want: %s, got: %s", ContentTypeApplicationJSON, gotEvent.Context.GetDataContentType())
	}

	gotExtension, err := gotEvent.Context.GetExtension(eventTypeVersionExtensionKey)
	if err != nil {
		t.Errorf("eventtype extension is missing: %v", err)
	}
	if gotExtension != legacyEventVersion {
		t.Errorf("incorrect eventtype extension, want: %s, got: %s", legacyEventVersion, gotExtension)
	}
}
