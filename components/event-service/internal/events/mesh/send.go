package mesh

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

func SendEvent(context context.Context, event cloudevents.Event) (*api.SendEventResponse, error) {
	// figure out the response back to the client
	response := &api.SendEventResponse{}

	// send the CE to the HTTP adapter
	// at that point the config is already initialized when the Event Service app is started
	_, _, err := config.CloudEventClient.Send(context, event)
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
