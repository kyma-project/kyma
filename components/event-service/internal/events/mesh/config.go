package mesh

import (
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go"
)

type Configuration struct {
	Source           string
	CloudEventClient cloudevents.Client
}

// GetConfig returns an Event mesh configuration instance which is required by the Event publish and delivery flows.
func GetConfig(source string, eventMeshUrl string) (*Configuration, error) {
	ceClient, err := getCloudEventClient(eventMeshUrl)
	if err != nil {
		return nil, err
	}
	config := &Configuration{
		Source:           source,
		CloudEventClient: ceClient,
	}
	return config, nil
}

// getCloudEventClient initializes a cloudevent client which points to the HTTP adapter created via the "Event Source",
// this is the internal entrypoint to our Event mesh.
func getCloudEventClient(eventMeshUrl string) (cloudevents.Client, error) {
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
	client, err := cloudevents.NewClient(transport, cloudevents.WithTimeNow())
	if err != nil {
		return nil, err
	}
	return client, nil
}
