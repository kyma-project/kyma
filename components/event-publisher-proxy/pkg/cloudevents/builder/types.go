package builder

import (
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
)

const (
	OriginalTypeHeaderName = "originaltype"
)

type CloudEventBuilder interface {
	Build(event cev2event.Event) (*cev2event.Event, error)
}

type GenericBuilder struct {
	typePrefix        string
	applicationLister *application.Lister // applicationLister will be nil when disabled.
	cleaner           cleaner.Cleaner
	logger            *logger.Logger
}

type EventMeshBuilder struct {
	genericBuilder     *GenericBuilder
	eventMeshNamespace string
}
