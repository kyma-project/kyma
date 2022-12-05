package legacy

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/legacytest"
	. "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	eventTypeMultiSegment         = "Segment1.Segment2.Segment3.Segment4.Segment5"
	eventTypeMultiSegmentCombined = "Segment1Segment2Segment3Segment4.Segment5"
)

// TestTransformLegacyRequestsToCE ensures that TransformLegacyRequestsToCE transforms a http request containing
// a legacy request to a valid cloud event by creating mock http requests
func TestTransformLegacyRequestsToCE(t *testing.T) {
	testCases := []struct {
		name string
		// the event type has the structure
		// <PREFIX>.<APPNAME>.<EVENTNAME>.<VERSION>
		// or, more specific
		// <PREFIX-1>.<PREFIX-2>.<PREFIX-3>.<APPNAME>.<EVENTNAME-1>.<EVENTNAME-2>.<VERSION>
		// e.g. "sap.kyma.custom.varkestest.ordertest.created.v1"
		// where "sap.kyma.custom" is the <PREFIX> (or <PREFIX-1>.<PREFIX-2>.<PREFIX-3>),
		// "varkestest" is the <APPNAME>
		// "ordertest.created" ts the <EVENTNAME> (or <EVENTNAME-1>.<EVENTNAME-2>)
		// and "v2" is the <VERSION>
		//
		// derived from that a legacy event path (a.k.a. endpoint) has the structure
		// http://<HOST>:<PORT>/<APPNAME>/<VERSION>/events
		// e.g. http://localhost:8081/varkestest/v1/events
		givenPrefix      string
		givenApplication string
		givenTypeLabel   string
		givenEventName   string
		wantVersion      string
		wantType         string
	}{
		{
			name:             "clean",
			givenPrefix:      "pre1.pre2.pre3",
			givenApplication: "app",
			givenTypeLabel:   "",
			givenEventName:   "object.do",
			wantVersion:      "v1",
			wantType:         "pre1.pre2.pre3.app.object.do.v1",
		},
		{
			name:             "unclean",
			givenPrefix:      "p!r@e&1.p,r:e2.p|r+e3",
			givenApplication: "m(i_s+h*a}p",
			givenTypeLabel:   "",
			givenEventName:   "o{b?j>e$c't.d;o",
			wantVersion:      "v1",
			wantType:         "pre1.pre2.pre3.mishap.object.do.v1",
		},
		{
			name:             "event name too many segments",
			givenPrefix:      "pre1.pre2.pre3",
			givenApplication: "app",
			givenTypeLabel:   "",
			givenEventName:   "too.many.dots.object.do",
			wantVersion:      "v1",
			wantType:         "pre1.pre2.pre3.app.toomanydotsobject.do.v1",
		},
		{
			name:             "with event type label",
			givenPrefix:      "pre1.pre2.pre3",
			givenApplication: "app",
			givenTypeLabel:   "different",
			givenEventName:   "object.do",
			wantVersion:      "v1",
			wantType:         "pre1.pre2.pre3.different.object.do.v1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			request, err := legacytest.ValidLegacyRequest(tc.wantVersion, tc.givenApplication, tc.givenEventName)
			assert.NoError(t, err)

			writer := httptest.NewRecorder()
			app := applicationtest.NewApplication(tc.givenApplication, applicationTypeLabel(tc.givenTypeLabel))
			appLister := fake.NewApplicationListerOrDie(ctx, app)
			transformer := NewTransformer("test", tc.givenPrefix, appLister)

			gotEvent, gotOriginalEventType := transformer.TransformLegacyRequestsToTransitionEvent(writer, request)
			wantEventType := formatEventType(tc.givenPrefix, tc.givenApplication, tc.givenEventName, tc.wantVersion)
			assert.Equal(t, wantEventType, gotOriginalEventType)

			// Check eventType.
			gotType := gotEvent.Context.GetType()
			assert.Equal(t, tc.wantType, gotType)

			// Check extensions 'eventtypeversion'.
			gotVersion, ok := gotEvent.Extensions()["eventtypeversion"].(string)
			assert.True(t, ok)
			assert.Equal(t, tc.wantVersion, gotVersion)

			// Check HTTP ContentType set properly.
			gotContentType := gotEvent.Context.GetDataContentType()
			assert.Equal(t, internal.ContentTypeApplicationJSON, gotContentType)
		})
	}
}

func applicationTypeLabel(label string) map[string]string {
	if label != "" {
		return map[string]string{application.TypeLabel: label}
	}
	return nil
}

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	givenEventID := EventID
	givenApplicationName := ApplicationName
	givenEventTypePrefix := Prefix
	givenTimeNow := time.Now().Format(time.RFC3339)
	givenLegacyEventVersion := EventVersion
	givenPublishReqParams := &legacyapi.PublishEventParametersV1{
		PublishrequestV1: legacyapi.PublishRequestV1{
			EventID:          givenEventID,
			EventType:        eventTypeMultiSegment,
			EventTime:        givenTimeNow,
			EventTypeVersion: givenLegacyEventVersion,
			Data:             EventData,
		},
	}

	wantBEBNamespace := MessagingNamespace
	wantEventID := givenEventID
	wantEventType := formatEventType(givenEventTypePrefix, givenApplicationName, eventTypeMultiSegmentCombined, givenLegacyEventVersion)
	wantTimeNowFormatted, _ := time.Parse(time.RFC3339, givenTimeNow)
	wantDataContentType := internal.ContentTypeApplicationJSON

	legacyTransformer := NewTransformer(wantBEBNamespace, givenEventTypePrefix, nil)
	gotEvent, err := legacyTransformer.convertPublishRequestToTransitionEvent(givenApplicationName, givenPublishReqParams)
	require.NoError(t, err)
	assert.Equal(t, wantBEBNamespace, gotEvent.Context.GetSource())
	assert.Equal(t, wantEventID, gotEvent.Context.GetID())
	assert.Equal(t, wantEventType, gotEvent.Context.GetType())
	assert.Equal(t, wantTimeNowFormatted, gotEvent.Context.GetTime())
	assert.Equal(t, wantDataContentType, gotEvent.Context.GetDataContentType())

	wantLegacyEventVersion := givenLegacyEventVersion
	gotExtension, err := gotEvent.Context.GetExtension(eventTypeVersionExtensionKey)
	assert.NoError(t, err)
	assert.Equal(t, wantLegacyEventVersion, gotExtension)
}

func TestCombineEventTypeSegments(t *testing.T) {
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
	}{
		{
			name:           "event-type with two segments",
			givenEventType: EventName,
			wantEventType:  EventName,
		},
		{
			name:           "event-type with more than two segments",
			givenEventType: eventTypeMultiSegment,
			wantEventType:  eventTypeMultiSegmentCombined,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if gotEventType := combineEventNameSegments(tc.givenEventType); tc.wantEventType != gotEventType {
				t.Fatalf("invalid event-type want: %s, got: %s", tc.wantEventType, gotEventType)
			}
		})
	}
}

func TestRemoveNonAlphanumeric(t *testing.T) {
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
	}{
		{
			name:           "unclean",
			givenEventType: "1-2+3=4.t&h$i#s.t!h@a%t.t;o/m$f*o{o]lery",
			wantEventType:  "1234.this.that.tomfoolery",
		},
		{
			name:           "clean",
			givenEventType: "1234.this.that",
			wantEventType:  "1234.this.that",
		},
		{
			name:           "single unclean segment",
			givenEventType: "t_o_m_f_o_o_l_e_r_y",
			wantEventType:  "tomfoolery",
		},
		{
			name:           "empty",
			givenEventType: "",
			wantEventType:  "",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s eventType", tc.name), func(t *testing.T) {
			t.Parallel()

			gotEventType := removeNonAlphanumeric(tc.givenEventType)
			assert.Equal(t, tc.wantEventType, gotEventType)
		})
	}
}
