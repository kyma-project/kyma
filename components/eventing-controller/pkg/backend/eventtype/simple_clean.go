package eventtype

import (
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
)

const (
	// simpleCleanerName used as the logger name.
	simpleCleanerName = "event-type-simple-cleaner"
)

type simpleCleaner struct {
	eventTypePrefix string
	logger          *logger.Logger
}

// Perform a compile-time check.
var _ Cleaner = &simpleCleaner{}

func NewSimpleCleaner(eventTypePrefix string, logger *logger.Logger) Cleaner {
	return &simpleCleaner{eventTypePrefix: eventTypePrefix, logger: logger}
}

// Clean cleans the event-type from none-alphanumeric characters and returns it
// or returns an error if the event-type parsing failed. It does not search for existing applications.
func (sc *simpleCleaner) Clean(eventType string) (string, error) {
	// format logger
	log := sc.namedLogger().With("prefix", sc.eventTypePrefix, "type", eventType)

	appName, event, version, err := parse(eventType, sc.eventTypePrefix)
	if err != nil {
		return "", err
	}

	// clean the application name
	var eventTypeClean = build(sc.eventTypePrefix, application.GetCleanName(appName), event, version)

	// clean the event-type segments
	eventTypeClean = cleanEventType(eventTypeClean)
	log.Debugw("Cleaned event-type",
		"before", eventType,
		"after", eventTypeClean,
	)

	return eventTypeClean, nil
}

func (sc *simpleCleaner) namedLogger() *zap.SugaredLogger {
	return sc.logger.WithContext().Named(simpleCleanerName)
}
