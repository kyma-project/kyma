package eventtype

import (
	"regexp"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
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

// CleanerFunc implements the Cleaner interface.
type CleanerFunc func(et string) (string, error)

func (cf CleanerFunc) Clean(et string) (string, error) {
	return cf(et)
}

type cleaner struct {
	eventTypePrefix string
	logger          *logger.Logger
}

// Perform a compile-time check.
var _ Cleaner = &cleaner{}

func NewCleaner(eventTypePrefix string, logger *logger.Logger) Cleaner {
	return &cleaner{eventTypePrefix: eventTypePrefix, logger: logger}
}

// Clean cleans the event-type from none-alphanumeric characters and returns it
// or returns an error if the event-type parsing failed.
func (c *cleaner) Clean(eventType string) (string, error) {
	// format logger
	log := c.namedLogger().With("prefix", c.eventTypePrefix, "type", eventType)

	appName, event, version, err := parse(eventType, c.eventTypePrefix)
	if err != nil {
		return "", err
	}

	// clean the application name
	eventTypeClean := build(c.eventTypePrefix, getCleanName(appName), event, version)

	// clean the event-type segments
	eventTypeClean = cleanEventType(eventTypeClean)
	log.Debugw("Cleaned event-type",
		"before", eventType,
		"after", eventTypeClean,
	)

	return eventTypeClean, nil
}

func (c *cleaner) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(cleanerName)
}

func cleanEventType(eventType string) string {
	return invalidEventTypeSegment.ReplaceAllString(eventType, "")
}
