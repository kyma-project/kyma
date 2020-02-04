package mesh

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

func SendEvent(context context.Context, event cloudevents.Event) (*api.SendEventResponse, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	ctx, evt, err := config.CloudEventClient.Send(context, event)
	if err != nil {
		return nil, err
	}
}
