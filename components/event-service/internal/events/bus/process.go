package bus

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
)

type configurationData struct {
	SourceID string
}

//Conf Event-Service configuration data
var Conf *configurationData

var eventsTargetURLV1 string
var eventsTargetURLV2 string
var clientV2 *cloudevents.Client

// Init should be used to initialize the "source" related configuration data
func Init(sourceID string, targetURLV1 string, targetURLV2 string) error {
	Conf = &configurationData{
		SourceID: sourceID,
	}
	eventsTargetURLV1 = targetURLV1
	eventsTargetURLV2 = targetURLV2

	// init cloud events client
	_clientV2, err := newClient()
	if err != nil {
		return err
	}
	clientV2 = _clientV2
	return nil
}

//CheckConf assert the configuration initialization
func CheckConf() (err error) {
	if Conf == nil {
		return fmt.Errorf("configuration data not initialized")
	}
	return nil
}

func newClient() (*cloudevents.Client, error) {
	options := []cloudeventshttp.Option{
		cloudevents.WithTarget(eventsTargetURLV2),
		cloudevents.WithStructuredEncoding(),
	}

	t, err := cloudevents.NewHTTPTransport(options...)

	if err != nil {
		return nil, err
	}

	c, err := cloudevents.NewClient(t,
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

