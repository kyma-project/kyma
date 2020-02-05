package mesh

import (
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	eventType        = "test-type"
	eventTypeVersion = "v1"
	eventTime        = "2018-11-02T22:08:41+00:00"
	data             = `{"data": "somejson"}`
	meshUrl          = "http://localhost:8080"
)

func TestConvertPublishRequestToCloudEvent(t *testing.T) {
	if err := Init("mocks", meshUrl); err != nil {
		t.Fatal(err)
	}

	publishRequest := &apiv1.PublishEventParametersV1{PublishrequestV1: apiv1.PublishRequestV1{
		EventType:        eventType,
		EventTypeVersion: eventTypeVersion,
		EventTime:        eventTime,
		Data:             data,
	}}

	cloudEvent, err := convertPublishRequestToCloudEvent(publishRequest)
	if err != nil {
		t.Fatalf("error occourred while converting publish request to cloudevent %+v", err)
	}

	log.Debugf("cloudevent object: %+v", cloudEvent)

	assert.Equal(t, cloudEvent.Type(), eventType)
	assert.Equal(t, cloudEvent.Extensions()["eventtypeversion"], eventTypeVersion)
	assert.NotNil(t, cloudEvent.Source())
	assert.NotNil(t, cloudEvent.ID())
}
