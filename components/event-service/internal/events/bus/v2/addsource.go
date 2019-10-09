package v2

import (
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
)

// AddSource adds the "source" related data to the incoming request
func AddSource(event cloudevents.Event) (*cloudevents.Event, error) {
	if err := bus.CheckConf(); err != nil {
		return nil, err
	}
	event.SetSource(bus.Conf.SourceID)
	return &event, nil
}
