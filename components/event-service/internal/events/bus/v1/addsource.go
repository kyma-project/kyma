package v1

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
)

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParametersV1) (resp *api.SendEventParametersV1, err error) {
	if err := bus.CheckConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParametersV1{
		SourceID:         bus.Conf.SourceID, // enrich the event with the sourceID
		EventType:        parameters.PublishrequestV1.EventType,
		EventTypeVersion: parameters.PublishrequestV1.EventTypeVersion,
		EventID:          parameters.PublishrequestV1.EventID,
		EventTime:        parameters.PublishrequestV1.EventTime,
		Data:             parameters.PublishrequestV1.Data,
	}

	return &sendRequest, nil
}
