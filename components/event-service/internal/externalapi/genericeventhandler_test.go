package externalapi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

// TODO(marcobebway) write tests

func TestCheckParameters(t *testing.T) {
	t.Parallel()

	// test meta
	const (
		eventType        = "test-type"
		eventTypeVersion = "v1"
		eventTime        = "2018-11-02T22:08:41+00:00"
		eventID          = "8954ad1c-78ed-4c58-a639-68bd44031de0"
		data             = `{"data": "somejson"}`
		dataEmpty        = ""
		invalid          = "!"
	)

	// test cases
	tests := []struct {
		name  string
		given *apiv1.PublishEventParametersV1
		want  *api.PublishEventResponses
	}{
		{
			name:  "nil params",
			given: nil,
			want:  shared.ErrorResponseBadRequest(shared.ErrorMessageBadPayload),
		},
		{
			name:  "missing field event type",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{}},
			want:  shared.ErrorResponseMissingFieldEventType(),
		},
		{
			name: "missing field event type version",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType: eventType,
			}},
			want: shared.ErrorResponseMissingFieldEventTypeVersion(),
		},
		{
			name: "wrong event type version",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: invalid,
			}},
			want: shared.ErrorResponseWrongEventTypeVersion(),
		},
		{
			name: "missing field event time",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
			}},
			want: shared.ErrorResponseMissingFieldEventTime(),
		},
		{
			name: "wrong event time",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        invalid,
			}},
			want: shared.ErrorResponseWrongEventTime(),
		},
		{
			name: "wrong event id",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        eventTime,
				EventID:          invalid,
			}},
			want: shared.ErrorResponseWrongEventID(),
		},
		{
			name: "missing field data",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        eventTime,
				EventID:          eventID,
			}},
			want: shared.ErrorResponseMissingFieldData(),
		},
		{
			name: "empty field data",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        eventTime,
				EventID:          eventID,
				Data:             dataEmpty,
			}},
			want: shared.ErrorResponseMissingFieldData(),
		},
		{
			name: "success",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        eventTime,
				EventID:          eventID,
				Data:             data,
			}},
			want: &api.PublishEventResponses{},
		},
	}

	// run all tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := checkParameters(test.given)
			if diff := cmp.Diff(got, test.want); len(diff) > 0 {
				t.Errorf("test '%s' failed:\n%s", test.name, diff)
			}
		})
	}
}
