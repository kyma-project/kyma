package eventtype

import (
	"regexp"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
)

var (
	// invalidEventTypeSegment used to match and replace none-alphanumeric characters in the event-type segments
	// as per SAP Event spec https://github.tools.sap/CentralEngineering/sap-event-specification#type.
	invalidEventTypeSegment = regexp.MustCompile("[^a-zA-Z0-9.]")

	// cleanerName used as the logger name.
	cleanerName = "event-type-cleaner"
)

type Cleaner interface {
	Clean(eventType string) (string, error)
}

type CleanerFunc func(et string) (string, error)

func (cf CleanerFunc) Clean(et string) (string, error) {
	return cf(et)
}

type cleaner struct {
	eventTypePrefix   string
	applicationLister *application.Lister
	logger            *logger.Logger
}

// compile-time check
var _ Cleaner = &cleaner{}

func NewCleaner(eventTypePrefix string, applicationLister *application.Lister, logger *logger.Logger) Cleaner {
	return &cleaner{eventTypePrefix: eventTypePrefix, applicationLister: applicationLister, logger: logger}
}

// Clean cleans the event-type from none-alphanumeric characters and returns it
// or returns an error if the event-type parsing failed
func (c *cleaner) Clean(eventType string) (string, error) {
	// format logger
	log := c.namedLogger().With("prefix", c.eventTypePrefix, "type", eventType)

	appName, event, version, err := parse(eventType, c.eventTypePrefix)
	if err != nil {
		log.Errorw("parse event-type failed", "error", err)
		return "", err
	}

	// clean the application name
	var eventTypeClean string
	if appObj, err := c.applicationLister.Get(appName); err != nil {
		log.Debugw("cannot find application", "name", appName)
		eventTypeClean = build(c.eventTypePrefix, application.GetCleanName(appName), event, version)
	} else {
		eventTypeClean = build(c.eventTypePrefix, application.GetCleanTypeOrName(appObj), event, version)
	}

	// clean the event-type segments
	eventTypeClean = cleanEventType(eventTypeClean)
	log.Debugw("clean event-type", "before", eventType, "after", eventTypeClean)

	return eventTypeClean, nil
}

func (c *cleaner) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(cleanerName)
}

func cleanEventType(eventType string) string {
	return invalidEventTypeSegment.ReplaceAllString(eventType, "")
}
