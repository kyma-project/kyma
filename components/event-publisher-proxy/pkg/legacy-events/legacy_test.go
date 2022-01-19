package legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	. "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	eventTypeMultiSegment         = "Segment1.Segment2.Segment3.Segment4.Segment5"
	eventTypeMultiSegmentCombined = "Segment1Segment2Segment3Segment4.Segment5"
)

// TestTransformLegacyRequestsToCE ensures that TransformLegacyRequestsToCE transforms a http reuqest containing
// a legacy request to a valid cloud event by creating mock http requests
func TestTransformLegacyRequestsToCE(t *testing.T) {
	testCases := []struct {
		name string
		// the event type has the strurcture
		// <PREFIX>.<APPNAME>.<EVENTNAME>.<VERSION>
		// or, more specific
		// <PREFIX-1>.<PREFIX-2>.<PREFIX-3>.<APPNAME>.<EVENTNAME-1>.<EVENTNAME-2>.<VERSION>
		// e.g. "sap.kyma.custom.varkestest.ordertest.created.v1"
		// where "sap.kyma.custom" is the <PREFIX> (or <PREFIX-1>.<PREFIX-2>.<PREFIX-3>),
		// "varkestest" is the <APPNAME>
		// "ordertest.created" ts the <EVENTNAME> (or <EVENTNAME-1>.<EVENTNAME-2>)
		// and "v2" is the <VERSION>
		//
		// derrived from that a legacy event path (a.k.a. endpoint) has the structure
		// http://<HOST>:<PORT>/<APPNAME>/<VERSION>/events
		// e.g. http://localhost:8081/varkestest/v1/events
		prefix       string
		appName      string
		typeLabel    string
		eventName    string
		version      string
		expectedType string
	}{
		{
			name:         "clean",
			prefix:       "pre1.pre2.pre3",
			appName:      "app",
			typeLabel:    "",
			eventName:    "object.do",
			version:      "v1",
			expectedType: "pre1.pre2.pre3.app.object.do.v1",
		},
		{
			name:         "unclean",
			prefix:       "p!r@e&1.p,r:e2.p|r+e3",
			appName:      "m(i_s+h*a}p",
			typeLabel:    "",
			eventName:    "o{b?j>e$c't.d;o",
			version:      "v1",
			expectedType: "pre1.pre2.pre3.mishap.object.do.v1",
		},
		{
			name:         "event name too many segments",
			prefix:       "pre1.pre2.pre3",
			appName:      "app",
			typeLabel:    "",
			eventName:    "too.many.dots.object.do",
			version:      "v1",
			expectedType: "pre1.pre2.pre3.app.toomanydotsobject.do.v1",
		},
		{
			name:         "with event type label",
			prefix:       "pre1.pre2.pre3",
			appName:      "app",
			typeLabel:    "different",
			eventName:    "object.do",
			version:      "v1",
			expectedType: "pre1.pre2.pre3.different.object.do.v1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			app := applicationtest.NewApplication(tc.appName, applicationTypeLabel(tc.typeLabel))
			appLister := fake.NewListerOrDie(ctx, app)

			request, err := mockLegacyRequest(tc.version, tc.appName, tc.eventName)
			if err != nil {
				t.Errorf("error while creating mock legacy request in case %s:\n%s", tc.name, err.Error())
			}

			writer := httptest.NewRecorder()

			transformer := NewTransformer("test", tc.prefix, appLister)
			actualEvent := transformer.TransformLegacyRequestsToCE(writer, request)

			errorString := func(caseName, subject, actual, expected string) string {
				return fmt.Sprintf("Unexpected %s in case %s\nwas:\n\t%s\nexpected:\n\t%s", subject, caseName, actual, expected)
			}

			//check eventType
			if actual := actualEvent.Context.GetType(); tc.expectedType != actual {
				t.Errorf(errorString(tc.name, "event type", actual, tc.expectedType))
			}

			// check extensions 'eventtypeversion'
			if actual := actualEvent.Extensions()["eventtypeversion"].(string); actual != tc.version {
				t.Errorf(errorString(tc.name, "event type", actual, tc.expectedType))
			}

			// check HTTP ContentType set properly
			if actualContentType := actualEvent.Context.GetDataContentType(); actualContentType != ContentTypeApplicationJSON {
				t.Errorf(errorString(tc.name, "HTTP Content Type", actualContentType, ContentTypeApplicationJSON))
			}
		})
	}
}

func applicationTypeLabel(label string) map[string]string {
	if label != "" {
		return map[string]string{"application-type": label}
	}
	return nil
}

func mockLegacyRequest(version, appname, eventType string) (*http.Request, error) {
	body, err := json.Marshal(map[string]string{
		"event-type":         eventType,
		"event-type-version": version,
		"event-time":         "2020-04-02T21:37:00Z",
		"data":               "{\"legacy\":\"event\"}",
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://localhost:8080/%s/%s/events", appname, version)
	return http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
}

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	bebNs := MessagingNamespace
	eventTypePrefix := MessagingEventTypePrefix
	legacyTransformer := NewTransformer(bebNs, eventTypePrefix, nil)
	eventID := EventID
	appName := ApplicationName
	legacyEventVersion := LegacyEventTypeVersion
	data := LegacyEventData
	timeNow := time.Now()
	expectedEventType := formatEventType4BEB(eventTypePrefix, appName, eventTypeMultiSegmentCombined, legacyEventVersion)
	timeNowStr := timeNow.Format(time.RFC3339)
	timeNowFormatted, _ := time.Parse(time.RFC3339, timeNowStr)
	publishReqParams := &legacyapi.PublishEventParametersV1{
		PublishrequestV1: legacyapi.PublishRequestV1{
			EventID:          eventID,
			EventType:        eventTypeMultiSegment,
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

func TestCombineEventTypeSegments(t *testing.T) {
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
	}{
		{
			name:           "event-type with two segments",
			givenEventType: LegacyEventType,
			wantEventType:  LegacyEventType,
		},
		{
			name:           "event-type with more than two segments",
			givenEventType: eventTypeMultiSegment,
			wantEventType:  eventTypeMultiSegmentCombined,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotEventType := combineEventNameSegments(tc.givenEventType); tc.wantEventType != gotEventType {
				t.Fatalf("invalid event-type want: %s, got: %s", tc.wantEventType, gotEventType)
			}
		})
	}
}

func TestRemoveNonAlphanumeric(t *testing.T) {
	testCases := []struct {
		name              string
		givenEventType    string
		expectedEventType string
	}{
		{
			name:              "unclean",
			givenEventType:    "1-2+3=4.t&h$i#s.t!h@a%t.t;o/m$f*o{o]lery",
			expectedEventType: "1234.this.that.tomfoolery",
		},
		{
			name:              "clean",
			givenEventType:    "1234.this.that",
			expectedEventType: "1234.this.that",
		},
		{
			name:              "single unclean segment",
			givenEventType:    "t_o_m_f_o_o_l_e_r_y",
			expectedEventType: "tomfoolery",
		},
		{
			name:              "empty",
			givenEventType:    "",
			expectedEventType: "",
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s eventType", tc.name), func(t *testing.T) {
			actual := removeNonAlphanumeric(tc.givenEventType)
			if actual != tc.expectedEventType {
				t.Errorf("invalid eventType; expected: %s, got %s", tc.expectedEventType, actual)
			}
		})
	}
}
