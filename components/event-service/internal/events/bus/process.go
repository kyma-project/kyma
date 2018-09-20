package bus

import (
	"fmt"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

type configurationData struct {
	sourceId string
}

var conf *configurationData
var eventsTargetURL string

// Init should be used to initialize the "source" related configuration data
func Init(sourceId string, targetUrl string) {
	conf = &configurationData{
		sourceId: sourceId,
	}

	eventsTargetURL = targetUrl
}

func checkConf() (err error) {
	if conf == nil {
		return fmt.Errorf("configuration data not initialized")
	}
	return nil
}

// AddSource adds the "source" related data to the incoming request
func AddSource(parameters *api.PublishEventParameters) (resp *api.SendEventParameters, err error) {
	if err := checkConf(); err != nil {
		return nil, err
	}

	sendRequest := api.SendEventParameters{
		SourceId:         conf.sourceId, // enrich the event with the sourceId
		EventType:        parameters.Publishrequest.EventType,
		EventTypeVersion: parameters.Publishrequest.EventTypeVersion,
		EventId:          parameters.Publishrequest.EventId,
		EventTime:        parameters.Publishrequest.EventTime,
		Data:             parameters.Publishrequest.Data,
	}

	return &sendRequest, nil
}
