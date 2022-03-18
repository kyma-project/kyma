package eventtype

import (
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
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

type cleaner struct {
	eventTypePrefix   string
	applicationLister *application.Lister
	logger            *logrus.Logger
}

// compile-time check
var _ Cleaner = &cleaner{}

func NewCleaner(eventTypePrefix string, applicationLister *application.Lister, logger *logrus.Logger) Cleaner {
	return &cleaner{eventTypePrefix: eventTypePrefix, applicationLister: applicationLister, logger: logger}
}

// Clean cleans the event-type from none-alphanumeric characters and returns it
// or returns an error if the event-type parsing failed.
func (c *cleaner) Clean(eventType string) (string, error) {
	// format logger
	log := c.namedLogger().WithFields(logrus.Fields{
		"prefix": c.eventTypePrefix,
		"type":   eventType,
	})

	appName, event, version, err := parse(eventType, c.eventTypePrefix)
	if err != nil {
		log.WithField("error", err).Error("parse event-type failed")
		return "", err
	}

	// clean the application name
	var eventTypeClean string
	if appObj, err := c.applicationLister.Get(appName); err != nil {
		log.WithField("application", appName).Debug("cannot find application")
		eventTypeClean = build(c.eventTypePrefix, application.GetCleanName(appName), event, version)
	} else {
		eventTypeClean = build(c.eventTypePrefix, application.GetCleanTypeOrName(appObj), event, version)
	}

	// clean the event-type segments
	eventTypeClean = cleanEventType(eventTypeClean)
	log.WithFields(
		logrus.Fields{
			"before": eventType,
			"after":  eventTypeClean,
		},
	).Debug("clean event-type")

	return eventTypeClean, nil
}

func (c *cleaner) namedLogger() *logrus.Logger {
	return c.logger.WithField("name", cleanerName).Logger
}

func cleanEventType(eventType string) string {
	return invalidEventTypeSegment.ReplaceAllString(eventType, "")
}
