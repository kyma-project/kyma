package v2

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/api/v2"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
)

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParametersV2) (resp *api.SendEventParametersV2, err error) {
	if err := bus.CheckConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParametersV2{
		Source:              bus.Conf.SourceID, // enrich the event with the sourceID
		Type:                parameters.EventRequestV2.EventType,
		EventTypeVersion:    parameters.EventRequestV2.EventTypeVersion,
		ID:                  parameters.EventRequestV2.EventID,
		Time:                parameters.EventRequestV2.EventTime,
		SpecVersion:         parameters.EventRequestV2.SpecVersion,
		DataContentEncoding: parameters.EventRequestV2.DataContentEncoding,
		Data:                parameters.EventRequestV2.Data,
	}

	return &sendRequest, nil
}
