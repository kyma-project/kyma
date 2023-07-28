package legacy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cev2event "github.com/cloudevents/sdk-go/v2/event"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/legacytest"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	eventTypeMultiSegment         = "Segment1.Segment2.Segment3.Segment4.Segment5"
	eventTypeMultiSegmentCombined = "Segment1Segment2Segment3Segment4.Segment5"
)

// TestTransformLegacyRequestsToCE ensures that TransformLegacyRequestsToCE transforms a http request containing
// a legacy request to a valid cloud event by creating mock http requests.
func TestTransformLegacyRequestsToCE(t *testing.T) {
	t.Parallel()
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
			name:             "not clean app name",
			givenPrefix:      "pre1.pre2.pre3",
			givenApplication: "no-app",
			givenTypeLabel:   "",
			givenEventName:   "object.do",
			wantVersion:      "v1",
			wantType:         "pre1.pre2.pre3.noapp.object.do.v1",
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
			publishData, errResp, _ := transformer.ExtractPublishRequestData(request)
			assert.Nil(t, errResp)
			gotEvent, gotEventType := transformer.WriteLegacyRequestsToCE(writer, publishData)
			wantEventType := formatEventType(tc.givenPrefix, tc.givenApplication, tc.givenEventName, tc.wantVersion)
			assert.Equal(t, wantEventType, gotEventType)

			// check eventType
			gotType := gotEvent.Context.GetType()
			assert.Equal(t, tc.wantType, gotType)

			// check extensions 'eventtypeversion'
			gotVersion, ok := gotEvent.Extensions()["eventtypeversion"].(string)
			assert.True(t, ok)
			assert.Equal(t, tc.wantVersion, gotVersion)

			// check HTTP ContentType set properly
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
	givenEventID := testingutils.EventID
	givenApplicationName := testingutils.ApplicationName
	givenEventTypePrefix := testingutils.Prefix
	givenTimeNow := time.Now().Format(time.RFC3339)
	givenLegacyEventVersion := testingutils.EventVersion
	givenPublishReqParams := &legacyapi.PublishEventParametersV1{
		PublishrequestV1: legacyapi.PublishRequestV1{
			EventID:          givenEventID,
			EventType:        eventTypeMultiSegment,
			EventTime:        givenTimeNow,
			EventTypeVersion: givenLegacyEventVersion,
			Data:             testingutils.EventData,
		},
	}

	wantEventMeshNamespace := testingutils.MessagingNamespace
	wantEventID := givenEventID
	wantEventType := formatEventType(givenEventTypePrefix, givenApplicationName, eventTypeMultiSegmentCombined, givenLegacyEventVersion)
	wantTimeNowFormatted, _ := time.Parse(time.RFC3339, givenTimeNow)
	wantDataContentType := internal.ContentTypeApplicationJSON

	legacyTransformer := NewTransformer(wantEventMeshNamespace, givenEventTypePrefix, nil)
	gotEvent, err := legacyTransformer.convertPublishRequestToCloudEvent(givenApplicationName, givenPublishReqParams)
	require.NoError(t, err)
	assert.Equal(t, wantEventMeshNamespace, gotEvent.Context.GetSource())
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
	t.Parallel()
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
	}{
		{
			name:           "event-type with two segments",
			givenEventType: testingutils.EventName,
			wantEventType:  testingutils.EventName,
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
	t.Parallel()
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

func TestExtractPublishRequestData(t *testing.T) {
	t.Parallel()
	const givenVersion = "v1"
	const givenPrefix = "pre1.pre2.pre3"
	const givenApplication = "app"
	const givenEventName = "object.do"

	testCases := []struct {
		name                   string
		givenLegacyRequestFunc func() (*http.Request, error)
		wantPublishRequestData *legacyapi.PublishRequestData
		wantErrorResponse      *legacyapi.PublishEventResponses
	}{
		{
			name: "should fail if request body is empty",
			givenLegacyRequestFunc: func() (*http.Request, error) {
				return legacytest.InvalidLegacyRequestWithEmptyBody(givenVersion, givenApplication)
			},
			wantErrorResponse: ErrorResponseBadRequest(ErrorMessageBadPayload),
		},
		{
			name: "should fail if request body has missing parameters",
			givenLegacyRequestFunc: func() (*http.Request, error) {
				return legacytest.InvalidLegacyRequest(givenVersion, givenApplication, givenEventName)
			},
			wantErrorResponse: ErrorResponseMissingFieldEventType(),
		},
		{
			name: "should succeed if request body is valid",
			givenLegacyRequestFunc: func() (*http.Request, error) {
				return legacytest.ValidLegacyRequest(givenVersion, givenApplication, givenEventName)
			},
			wantPublishRequestData: &legacyapi.PublishRequestData{
				PublishEventParameters: &legacyapi.PublishEventParametersV1{
					PublishrequestV1: legacyapi.PublishRequestV1{
						EventType:        "object.do",
						EventTypeVersion: "v1",
						EventTime:        "2020-04-02T21:37:00Z",
						Data:             "{\"legacy\":\"event\"}",
					},
				},
				ApplicationName: "app",
				URLPath:         "/app/v1/events",
			},
		},
		{
			name: "should succeed if app name has a special character",
			givenLegacyRequestFunc: func() (*http.Request, error) {
				return legacytest.ValidLegacyRequest(givenVersion, "no-app", givenEventName)
			},
			wantPublishRequestData: &legacyapi.PublishRequestData{
				PublishEventParameters: &legacyapi.PublishEventParametersV1{
					PublishrequestV1: legacyapi.PublishRequestV1{
						EventType:        "object.do",
						EventTypeVersion: "v1",
						EventTime:        "2020-04-02T21:37:00Z",
						Data:             "{\"legacy\":\"event\"}",
					},
				},
				ApplicationName: "no-app",
				URLPath:         "/no-app/v1/events",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			request, err := tc.givenLegacyRequestFunc()
			require.NoError(t, err)
			// set expected header
			if tc.wantPublishRequestData != nil {
				tc.wantPublishRequestData.Headers = request.Header
			}

			transformer := NewTransformer("test", givenPrefix, nil)

			// when
			publishData, errResp, err := transformer.ExtractPublishRequestData(request)

			// then
			if tc.wantErrorResponse != nil {
				require.Error(t, err)
				require.Equal(t, *tc.wantErrorResponse, *errResp)
			} else {
				require.NoError(t, err)
				require.Nil(t, errResp)
				require.Equal(t, *tc.wantPublishRequestData, *publishData)
			}
		})
	}
}

func TestTransformPublishRequestToCloudEvent(t *testing.T) {
	t.Parallel()
	const givenVersion = "v1"
	const givenPrefix = "pre1.pre2.pre3"
	const givenApplication = "app"
	const givenEventName = "object.do"

	testCases := []struct {
		name                        string
		givenPublishEventParameters legacyapi.PublishEventParametersV1
		wantCloudEventFunc          func() (cev2event.Event, error)
		wantErrorResponse           legacyapi.PublishEventResponses
		wantError                   bool
		wantEventType               string
		wantSource                  string
		wantData                    string
		wantEventID                 string
	}{
		{
			name: "should succeed if publish data is valid",
			givenPublishEventParameters: legacyapi.PublishEventParametersV1{
				PublishrequestV1: legacyapi.PublishRequestV1{
					EventID:          testingutils.EventID,
					EventType:        givenEventName,
					EventTypeVersion: givenVersion,
					EventTime:        "2020-04-02T21:37:00Z",
					Data:             map[string]string{"key": "value"},
				},
			},
			wantError:     false,
			wantEventID:   testingutils.EventID,
			wantEventType: "object.do.v1",
			wantSource:    givenApplication,
			wantData:      `{"key":"value"}`,
		},
		{
			name: "should set new event ID when not provided",
			givenPublishEventParameters: legacyapi.PublishEventParametersV1{
				PublishrequestV1: legacyapi.PublishRequestV1{
					EventType:        givenEventName,
					EventTypeVersion: givenVersion,
					EventTime:        "2020-04-02T21:37:00Z",
					Data:             map[string]string{"key": "value"},
				},
			},
			wantError:     false,
			wantEventType: "object.do.v1",
			wantSource:    givenApplication,
			wantData:      `{"key":"value"}`,
		},
		{
			name: "should fail if event time is invalid",
			givenPublishEventParameters: legacyapi.PublishEventParametersV1{
				PublishrequestV1: legacyapi.PublishRequestV1{
					EventType:        givenEventName,
					EventTypeVersion: givenVersion,
					EventTime:        "20dsadsa20-04-02T21:37:00Z",
					Data:             map[string]string{"key": "value"},
				},
			},
			wantError: true,
		},
		{
			name: "should fail if event data is not json",
			givenPublishEventParameters: legacyapi.PublishEventParametersV1{
				PublishrequestV1: legacyapi.PublishRequestV1{
					EventType:        givenEventName,
					EventTypeVersion: givenVersion,
					EventTime:        "20dsadsa20-04-02T21:37:00Z",
					Data:             "test",
				},
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			givenPublishRequestData := legacyapi.PublishRequestData{
				PublishEventParameters: &tc.givenPublishEventParameters,
				ApplicationName:        givenApplication,
			}
			transformer := NewTransformer("test", givenPrefix, nil)

			// when
			ceEvent, err := transformer.TransformPublishRequestToCloudEvent(&givenPublishRequestData)

			// then
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.Equal(t, tc.wantEventType, ceEvent.Type())
				require.Equal(t, tc.wantSource, ceEvent.Source())
				require.Equal(t, tc.wantData, string(ceEvent.Data()))
				require.NotEmpty(t, ceEvent.ID())
				if tc.wantEventID != "" {
					require.Equal(t, tc.wantEventID, ceEvent.ID())
				}
			}
		})
	}
}
