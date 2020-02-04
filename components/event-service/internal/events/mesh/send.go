package mesh

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	log "github.com/sirupsen/logrus"
	"time"
)

func SendEvent(context context.Context, publishRequest apiv1.PublishRequestV1) (*api.SendEventResponse, error) {
	// figure out the response back to the client
	response := &api.SendEventResponse{}

	evt, err := convertPublishRequestToCloudEvent(publishRequest)
	if err != nil {
		// TODO(marcobebway) figure this out
	}

	// send the CE to the HTTP adapter
	// at that point the config is already initialized when the Event Service app is started
	_, _, err = config.CloudEventClient.Send(context, *evt)
	if err != nil {
		// TODO(marcobebway) figure this out
		response.Error = &api.Error{
			Status:   0,
			Type:     "",
			Message:  "",
			MoreInfo: "",
			Details:  nil,
		}

		return response, err
	}

	// TODO(marcobebway) figure this out
	response.Ok = &api.PublishResponse{
		EventID: "",
		Status:  "",
		Reason:  "",
	}

	return response, nil
}

func convertPublishRequestToCloudEvent(publishRequest apiv1.PublishRequestV1) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	//Set "Source" value, which is currently name of the "Application"
	event.SetSource(config.Source)

	event.SetID(uuid.New().String())
	event.SetType(publishRequest.EventType)
	event.SetExtension("eventtypeversion", publishRequest.EventTypeVersion)

	t, err := time.Parse(time.RFC3339, publishRequest.EventTime)
	if err != nil {
		log.Errorf("error occurred in parsing time from the external publish request. Error Details:\n %+v", err)
	}
	event.SetTime(t)

	event.SetDataContentType(httpconsts.ContentTypeApplicationJSON)
	err = event.SetData(publishRequest.Data)
	if err != nil {
		log.Errorf("error occurred while setting data object. Error Details :\n %+v", err)
		return nil, err
	}
	return &event, nil
}
