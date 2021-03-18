package eventtype

import (
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
)

type Cleaner interface {
	Clean(eventType string) (string, error)
}

type cleaner struct {
	eventTypePrefix   string
	applicationLister *application.Lister
	logger            logr.Logger
}

// compile-time check
var _ Cleaner = &cleaner{}

func NewCleaner(eventTypePrefix string, applicationLister *application.Lister, logger logr.Logger) Cleaner {
	return &cleaner{eventTypePrefix: eventTypePrefix, applicationLister: applicationLister, logger: logger}
}

// Clean cleans the application name segment in the event-type from none-alphanumeric characters
// and returns the clean event-type, or returns an error if the event-type parsing failed
func (c cleaner) Clean(eventType string) (string, error) {
	appName, event, version, err := parse(eventType, c.eventTypePrefix)
	if err != nil {
		c.logger.Error(err, "failed to parse event-type", "prefix", c.eventTypePrefix, "type", eventType)
		return "", err
	}

	// handle existing applications
	if appObj, err := c.applicationLister.Get(appName); err == nil {
		eventType = build(c.eventTypePrefix, application.GetCleanTypeOrName(appObj), event, version)
		return eventType, nil
	}

	// handle non-existing applications
	c.logger.Info("failed to get application", "name", appName)
	eventType = build(c.eventTypePrefix, application.GetCleanName(appName), event, version)
	return eventType, nil
}
