package handlers

import (
	"fmt"
	"time"

	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func SendEventToNATS(natsClient *Nats, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := eventingtesting.EventType
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := NewNatsMessagePayload(data, "id", eventingtesting.EventSource, eventTime, eventType)
	return natsClient.connection.Publish(eventType, []byte(sampleEvent))
}

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}
