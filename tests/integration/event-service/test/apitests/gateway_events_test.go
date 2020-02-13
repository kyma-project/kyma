package apitests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/integration/event-service/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PublishRequest struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventId          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// AnyValue implements the service definition of AnyValue
type AnyValue interface {
}

type PublishResponse struct {
	EventId string `json:"event-id,omitempty"`
}

type SubscribedEvents struct {
	EventsInfo []Event `json:"eventsInfo"`
}

type Event struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

const (
	eventType = "test.eventtype"
)

func TestGatewayEvents(t *testing.T) {

	t.Run("should send events via Event Service", func(t *testing.T) {
		// given
		publishRequest := PublishRequest{
			EventType:        "order.created",
			EventTypeVersion: "v1",
			EventId:          "31109198-4d69-4ae0-972d-76117f3748c8",
			EventTime:        "2012-11-01T22:08:41+00:00",
			Data:             "payload",
		}
		publishRequestEncoded, err := json.Marshal(publishRequest)
		require.NoError(t, err)

		url := config.EventServiceUrl + "/" + config.Application + "/v1/events"

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(publishRequestEncoded))
		require.NoError(t, err)

		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		// when
		response, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		// then
		var publishResponse PublishResponse
		err = json.NewDecoder(response.Body).Decode(&publishResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "31109198-4d69-4ae0-972d-76117f3748c8", publishResponse.EventId)
	})

	t.Run("should get all subscribed events", func(t *testing.T) {
		//given
		client, e := testkit.NewTriggerClient()
		require.NoError(t, e)

		err := client.Create(config.Namespace, config.Application, eventType)
		require.NoError(t, err)

		url := config.EventServiceUrl + "/" + config.Application + "/v1/events/subscribed"

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		//when
		response, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		//then
		var events SubscribedEvents

		err = json.NewDecoder(response.Body).Decode(&events)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.True(t, containsEventName(events.EventsInfo, eventType))

		//cleanup
		err = client.Delete(config.Namespace)
		require.NoError(t, e)
	})
}

func containsEventName(events []Event, name string) bool {
	for _, e := range events {
		if e.Name == name {
			return true
		}
	}
	return false
}
