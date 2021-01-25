package legacy

import (
	"testing"
	"time"

	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
)

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	bebNs := "/beb.namespace"
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
