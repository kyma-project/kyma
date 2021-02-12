package testing

import (
	"net/url"
	"testing"
	"time"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

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
