package mesh

import (
	"context"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	log "github.com/sirupsen/logrus"
)

// SendEvent sends a CloudEvent to the application's HTTP Adapter using the CloudEvent client.
func SendEvent(config *Configuration, context context.Context, publishRequest *apiv1.PublishEventParametersV1) (*api.PublishEventResponses, error) {
	response := &api.PublishEventResponses{}

	evt, err := convertPublishRequestToCloudEvent(config, publishRequest)
	if err != nil {
		response.Error = &api.Error{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
		return response, err
	}

	rctx, _, err := config.CloudEventClient.Send(context, *evt)
	rtctx := cloudevents.HTTPTransportContextFrom(rctx)

	if err != nil {
		response.Error = &api.Error{
			Status:  rtctx.StatusCode,
			Message: err.Error(),
		}
		return response, err
	}

	if !is2XXStatusCode(rtctx.StatusCode) {
		response.Error = &api.Error{Status: rtctx.StatusCode}
		return response, nil
	}

	response.Ok = &api.PublishResponse{EventID: evt.ID()}
	return response, nil
}

// convertPublishRequestToCloudEvent converts the given publish request to a CloudEvent.
func convertPublishRequestToCloudEvent(config *Configuration, publishRequest *apiv1.PublishEventParametersV1) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	t, err := time.Parse(time.RFC3339, publishRequest.PublishrequestV1.EventTime)
	if err != nil {
		log.Errorf("error occurred in parsing time from the external publish request. Error Details:\n %+v", err)
		return nil, err
	}
	event.SetTime(t)

	if err := event.SetData(publishRequest.PublishrequestV1.Data); err != nil {
		log.Errorf("error occurred while setting data object. Error Details :\n %+v", err)
		return nil, err
	}

	// set the event id from the request if it is available
	// otherwise generate a new one
	if len(publishRequest.PublishrequestV1.EventID) > 0 {
		event.SetID(publishRequest.PublishrequestV1.EventID)
	} else {
		event.SetID(uuid.New().String())
	}

	event.SetSource(config.Source)
	event.SetType(publishRequest.PublishrequestV1.EventType)
	event.SetDataContentType(httpconsts.ContentTypeApplicationJSON)
	event.SetExtension("sourceid", config.Source)
	event.SetExtension("eventtypeversion", publishRequest.PublishrequestV1.EventTypeVersion)

	return &event, nil
}

// is2XXStatusCode checks whether status code is a 2XX status code
func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
