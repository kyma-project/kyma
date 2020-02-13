package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	source           = "mock"
	eventID          = "8954ad1c-78ed-4c58-a639-68bd44031de0"
	eventType        = "test-type"
	eventTypeVersion = "v1"
	eventTime        = "2018-11-02T22:08:41+00:00"
	data             = `{"data": "somejson"}`
	meshUrl          = "http://localhost:8080"
)

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	config, err := GetConfig(source, meshUrl)
	if err != nil {
		t.Fatal(err)
	}

	publishRequest := &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
		EventType:        eventType,
		EventID:          eventID,
		EventTypeVersion: eventTypeVersion,
		EventTime:        eventTime,
		Data:             data,
	}}

	cloudEvent, err := convertPublishRequestToCloudEvent(config, publishRequest)
	if err != nil {
		t.Fatalf("error occourred while converting publish request to cloudevent %+v", err)
	}

	log.Debugf("cloudevent object: %+v", cloudEvent)

	assert.Equal(t, cloudEvent.Type(), eventType)
	assert.Equal(t, cloudEvent.Extensions()["eventtypeversion"], eventTypeVersion)
	assert.NotEmpty(t, cloudEvent.Source())
	assert.Equal(t, cloudEvent.Source(), source)
	assert.NotEmpty(t, cloudEvent.ID())
}

func TestSendEvent(t *testing.T) {
	mockURL := mockEventMesh(t)

	config, err := GetConfig(source, *mockURL)
	if err != nil {
		t.Fatal(err)
	}

	publishRequest := &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
		EventID:          eventID,
		EventType:        eventType,
		EventTypeVersion: eventTypeVersion,
		EventTime:        eventTime,
		Data:             data,
	}}

	cloudEvent, err := convertPublishRequestToCloudEvent(config, publishRequest)
	if err != nil {
		t.Fatalf("error occourred while converting publish request to cloudevent %+v", err)
	}

	assert.Equal(t, cloudEvent.Type(), eventType)
	assert.Equal(t, cloudEvent.Extensions()["sourceid"], source)
	assert.Equal(t, cloudEvent.Extensions()["eventtypeversion"], eventTypeVersion)
	assert.NotEmpty(t, cloudEvent.Source())
	assert.Equal(t, cloudEvent.Source(), source)
	assert.NotEmpty(t, cloudEvent.ID())
	assert.Equal(t, cloudEvent.ID(), eventID)

	res, err := SendEvent(config, context.TODO(), publishRequest)

	assert.Nil(t, err)
	assert.Nil(t, res.Error)
	assert.NotEmpty(t, res.Ok)
	assert.Equal(t, eventID, res.Ok.EventID)
}

func mockEventMesh(t *testing.T) *string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write success response with a valid event id
		resp := &api.PublishEventResponses{Ok: &api.PublishResponse{EventID: r.Header.Get("ce-id")}}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp.Ok); err != nil {
			t.Fatalf("failed to write response")
		}
	}))

	if srv == nil {
		t.Fatalf("failed to start HTTP server")
		return nil
	}

	return &srv.URL
}
