package mesh

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// SendEvent TODO(marcobebway)
func SendEvent(context context.Context, publishRequest *apiv1.PublishEventParametersV1) (*api.PublishEventResponses, error) {
	// prepare the response
	response := &api.PublishEventResponses{}

	// convert the received event to a cloudevent
	evt, err := convertPublishRequestToCloudEvent(publishRequest)
	if err != nil {
		response.Error = &api.Error{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
		return response, err
	}

	// send the cloudevent to the HTTP adapter
	// at this point the config is already initialized when the Event Service app is started
	rctx, _, err := config.CloudEventClient.Send(context, *evt)
	rtctx := cloudevents.HTTPTransportContextFrom(rctx)

	// handle errors returned from the HTTP adapter
	if err != nil {
		response.Error = &api.Error{
			Status:  rtctx.StatusCode,
			Message: err.Error(),
		}
		return response, err
	}

	// accept only 2XX status code
	if !is2XXStatusCode(rtctx.StatusCode) {
		response.Error = &api.Error{Status: rtctx.StatusCode}
		return response, nil
	}

	// request is successful, send the response back
	response.Ok = &api.PublishResponse{EventID: evt.ID()}
	return response, nil
}

// convertPublishRequestToCloudEvent TODO(marcobebway)
func convertPublishRequestToCloudEvent(publishRequest *apiv1.PublishEventParametersV1) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	// set the event time
	if t, err := time.Parse(time.RFC3339, publishRequest.PublishrequestV1.EventTime); err != nil {
		log.Errorf("error occurred in parsing time from the external publish request. Error Details:\n %+v", err)
		return nil, err
	} else {
		event.SetTime(t)
	}

	// set the event data
	if err := event.SetData(publishRequest.PublishrequestV1.Data); err != nil {
		log.Errorf("error occurred while setting data object. Error Details :\n %+v", err)
		return nil, err
	}

	event.SetID(publishRequest.PublishrequestV1.EventID)
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
