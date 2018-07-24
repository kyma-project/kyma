package bus

import (
	"fmt"

	"github.com/kyma-project/kyma/components/gateway/internal/events/api"
)

type configurationData struct {
	sourceNamespace   *string
	sourceType        *string
	sourceEnvironment *string
}

var conf *configurationData
var eventsTargetURL string

// Init should be used to initialize the "source" related configuration data
func Init(sourceNamespace *string, sourceType *string, sourceEnvironment *string, targetUrl *string) {
	conf = &configurationData{
		sourceNamespace:   sourceNamespace,
		sourceType:        sourceType,
		sourceEnvironment: sourceEnvironment,
	}

	eventsTargetURL = *targetUrl
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

	es := api.EventSource{SourceNamespace: *conf.sourceNamespace, SourceType: *conf.sourceType, SourceEnvironment: *conf.sourceEnvironment}

	sendRequest := api.SendEventParameters{
		Eventsource:      es,
		EventType:        parameters.Publishrequest.EventType,
		EventTypeVersion: parameters.Publishrequest.EventTypeVersion,
		EventId:          parameters.Publishrequest.EventId,
		EventTime:        parameters.Publishrequest.EventTime,
		Data:             parameters.Publishrequest.Data,
	}

	return &sendRequest, nil
}
