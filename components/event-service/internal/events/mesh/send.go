package mesh

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func SendEvent(context context.Context, publishRequest *apiv1.PublishEventParametersV1) (*api.PublishEventResponses, error) {
	// figure out the response back to the client
	response := &api.PublishEventResponses{}

	evt, err := convertPublishRequestToCloudEvent(publishRequest)
	if err != nil {
		response.Error = &api.Error{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}

		return response, err
	}

	// send the CE to the HTTP adapter
	// at that point the config is already initialized when the Event Service app is started
	rctx, revt, err := config.CloudEventClient.Send(context, *evt)

	// TODO(marcobebway) make sure at this point we always have a non-nil context returned
	rtctx := cloudevents.HTTPTransportContextFrom(rctx)

	if err != nil {
		response.Error = &api.Error{
			Status:  rtctx.StatusCode,
			Message: err.Error(),
		}

		log.Debugf("error: %v", err)
		return response, err
	}

	if rtctx.StatusCode != http.StatusOK {
		log.Debugf("context: %v", rtctx)
		response.Error = &api.Error{Status: rtctx.StatusCode}
		return response, nil
	}

	response.Ok = &api.PublishResponse{EventID: revt.ID()}
	return response, nil
}

/*
	convert PublishRequestV1 to CE
*/
func convertPublishRequestToCloudEvent(publishRequest *apiv1.PublishEventParametersV1) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	/*
		generate an event id if there is none
	*/

	if len(publishRequest.PublishrequestV1.EventID) == 0 {
		event.SetID(uuid.New().String())
	}
	event.SetType(publishRequest.PublishrequestV1.EventType)
	event.SetExtension("eventtypeversion", publishRequest.PublishrequestV1.EventTypeVersion)
	event.SetExtension("sourceid", config.Source)

	t, err := time.Parse(time.RFC3339, publishRequest.PublishrequestV1.EventTime)
	if err != nil {
		log.Errorf("error occurred in parsing time from the external publish request. Error Details:\n %+v", err)
		return nil, err
	}
	event.SetTime(t)

	event.SetDataContentType(httpconsts.ContentTypeApplicationJSON)
	err = event.SetData(publishRequest.PublishrequestV1.Data)
	if err != nil {
		log.Errorf("error occurred while setting data object. Error Details :\n %+v", err)
		return nil, err
	}
	return &event, nil
}
