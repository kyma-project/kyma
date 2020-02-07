package mesh

import (
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go"
)

type configurationData struct {
	Source           string
	CloudEventClient cloudevents.Client
}

var config *configurationData

func Init(source string, eventMeshUrl string) error {
	ceClient, err := getCloudEventClient(eventMeshUrl)
	if err != nil {
		return err
	}
	config = &configurationData{
		Source:           source,
		CloudEventClient: ceClient,
	}
	return nil
}

// Initialize a cloudevent client which points to the HTTP adapter created via the "Event Source", this is the internal
// entrypoint to our event-mesh
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
	client, err := cloudevents.NewClient(transport, cloudevents.WithUUIDs(), cloudevents.WithTimeNow())
	if err != nil {
		return nil, err
	}
	return client, nil
}
