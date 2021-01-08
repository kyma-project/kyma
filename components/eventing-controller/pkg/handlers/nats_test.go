package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/nats-io/nats.go"
)

func TestConvertMsgToCE(t *testing.T) {
	eventTime := time.Now().Format(time.RFC3339)
	testCases := []struct {
		name               string
		natsMsg            nats.Msg
		expectedCloudEvent cev2event.Event
		expectedErr        error
	}{
		{
			name: "data without quotes",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("foo-data", "id", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: NewCloudEvent("foo-data", "id", "foosource", eventTime, "fooeventtype", t),
			expectedErr:        nil,
		},
		{
			name: "data with quotes",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("\\\"foo-data\\\"", "id", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: NewCloudEvent("\\\"foo-data\\\"", "id", "foosource", eventTime, "fooeventtype", t),
			expectedErr:        nil,
		},
		{
			name: "natsMessage which is an invalid Cloud Event with empty id",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("foo-data", "", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: cev2event.New(cev2event.CloudEventsVersionV1),
			expectedErr:        errors.New("id: MUST be a non-empty string\n"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			gotCE, err := convertMsgToCE(&tc.natsMsg)
			if err != nil && tc.expectedErr == nil {
				t.Errorf("should not give error, got: %v", err)
				return
			}
			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("received nil error, expected: %v got: %v", tc.expectedErr, err)
					return
				}
				if tc.expectedErr.Error() != err.Error() {
					t.Errorf("received wrong error, expected: %v got: %v", tc.expectedErr, err)
				}
				return
			}
			if !(gotCE.Subject() == tc.expectedCloudEvent.Subject()) ||
				!(gotCE.ID() == tc.expectedCloudEvent.ID()) ||
				!(gotCE.DataContentType() == tc.expectedCloudEvent.DataContentType()) ||
				!(gotCE.Source() == tc.expectedCloudEvent.Source()) ||
				!(gotCE.Time().String() == tc.expectedCloudEvent.Time().String()) ||
				!(string(gotCE.Data()) == string(tc.expectedCloudEvent.Data())) {
				t.Errorf("received wrong cloudevent, expected: %v got: %v", tc.expectedCloudEvent, gotCE)
			}
		})
	}
}

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}

func NewCloudEvent(data, id, source, eventTime, eventType string, t *testing.T) cev2event.Event {
	timeInRFC3339, err := time.Parse(time.RFC3339, eventTime)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	dataContentType := "application/json"
	return cev2event.Event{
		Context: &cev2event.EventContextV1{
			Type: eventType,
			Source: types.URIRef{
				URL: url.URL{
					Path: source,
				},
			},
			ID:              id,
			DataContentType: &dataContentType,
			Time:            &types.Timestamp{Time: timeInRFC3339},
		},
		DataEncoded: []byte(data),
	}
}
