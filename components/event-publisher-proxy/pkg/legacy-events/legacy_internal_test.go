//go:build unit
// +build unit

package legacy

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	. "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	eventTypeMultiSegment         = "Segment1.Segment2.Segment3.Segment4.Segment5"
	eventTypeMultiSegmentCombined = "Segment1Segment2Segment3Segment4.Segment5"
)

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	// given
	givenEventID := EventID
	givenApplicationName := ApplicationName
	givenEventTypePrefix := MessagingEventTypePrefix
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

	wantBebNs := MessagingNamespace
	wantEventID := givenEventID
	wantEventType := fmt.Sprintf("%s.%s.%s.%s", givenEventTypePrefix, givenApplicationName, eventTypeMultiSegmentCombined, givenLegacyEventVersion)
	wantTimeNowFormatted, _ := time.Parse(time.RFC3339, givenTimeNow)
	wantDataContentType := internal.ContentTypeApplicationJSON

	// when
	legacyTransformer := NewTransformer(wantBebNs, givenEventTypePrefix, nil)
	gotEvent, err := legacyTransformer.convertPublishRequestToCloudEvent(givenApplicationName, givenPublishReqParams)

	// then
	require.NoError(t, err)
	assert.Equal(t, wantBebNs, gotEvent.Context.GetSource())
	assert.Equal(t, wantEventID, gotEvent.Context.GetID())
	assert.Equal(t, wantEventType, gotEvent.Context.GetType())
	assert.Equal(t, wantTimeNowFormatted, gotEvent.Context.GetTime())
	assert.Equal(t, wantDataContentType, gotEvent.Context.GetDataContentType())

	// when
	wantLegacyEventVersion := givenLegacyEventVersion
	gotExtension, err := gotEvent.Context.GetExtension(eventTypeVersionExtensionKey)

	// then
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
