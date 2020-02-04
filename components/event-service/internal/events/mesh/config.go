package mesh

import (
	"fmt"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go"
)

type configuratationData struct {
	Source           string
	CloudEventClient cloudevents.Client
}

var config *configuratationData

func Init(source string, eventMeshUrl string) error {
	ceClient, err := getCloudEventClient(eventMeshUrl)
	if err != nil {
		return err
	}
	config = &configuratationData{
		Source:           source,
		CloudEventClient: ceClient,
	}
	return nil
}

// Get Initialized configurartion Data for the "event-service".
func getConfig() (configData *configuratationData, err error) {
	if config == nil {
		return nil, fmt.Errorf("configuration data not initialized")
	}
	return config, nil
}

// Initialize a cloudevent client which points to the HTTP adapter created via the "Event Source", this is the internal
// entrypoint to our event-mesh
func getCloudEventClient(eventMeshUrl string) (ceClient cloudevents.Client, err error) {
	if _, err := url.Parse(eventMeshUrl); err != nil {
		return nil, err
	}

	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(eventMeshUrl),
		cloudevents.WithBinaryEncoding(),
	)
	if err != nil {
		return nil, err
	}
	client, err := cloudevents.NewClient(transport)
	if err != nil {
		return nil, err
	}
	return client, nil
}
