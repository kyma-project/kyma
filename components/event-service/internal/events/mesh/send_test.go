package mesh

import (
	"context"
	"testing"

	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	meshtesting "github.com/kyma-project/kyma/components/event-service/internal/testing"
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
	config, err := InitConfig(source, meshUrl)
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

	log.Printf("Cloudevent object: %+v", cloudEvent)

	assert.Equal(t, cloudEvent.Type(), eventType)
	assert.Equal(t, cloudEvent.Extensions()["eventtypeversion"], eventTypeVersion)
	assert.NotEmpty(t, cloudEvent.Source())
	assert.Equal(t, cloudEvent.Source(), source)
	assert.NotEmpty(t, cloudEvent.ID())
}

func TestSendEvent(t *testing.T) {
	// setup
	mockURL, closeFn := meshtesting.MockEventMesh(t)
	defer closeFn()

	// test cases
	tests := []struct {
		name  string
		given *apiv1.PublishEventParametersV1
		want  string
	}{
		{
			name: "request with event id",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventID:          eventID,
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        eventTime,
				Data:             data,
			}},
			want: eventID,
		},
		{
			name: "request without event id",
			given: &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
				EventType:        eventType,
				EventTypeVersion: eventTypeVersion,
				EventTime:        eventTime,
				Data:             data,
			}},
			want: "",
		},
	}

	config, err := InitConfig(source, mockURL)
	if err != nil {
		t.Fatalf("test setup failed with error: %v", err)
	}

	// run all tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			res, err := SendEvent(config, context.TODO(), test.given)
			if err != nil {
				t.Fatalf("test '%s' failed with error: %v", test.name, err)
			}
			if res.Error != nil {
				t.Fatalf("test '%s' failed with returned response error: %v", test.name, res.Error)
			}
			if res.Ok == nil {
				t.Fatalf("test '%s' failed with returned response not ok", test.name)
			}
			if len(test.given.PublishrequestV1.EventID) > 0 && test.want != res.Ok.EventID {
				t.Fatalf("test '%s' failed with error event id mismatch, want: '%v' but got: '%v'", test.name, test.want, res.Ok.EventID)
			}
		})
	}
}
