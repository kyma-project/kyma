package eventtype

import (
	"regexp"

	"golang.org/x/xerrors"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
)

var (
	// invalidEventTypeSegment used to match and replace none-alphanumeric characters in the event-type segments
	// as per SAP Event spec https://github.tools.sap/CentralEngineering/sap-event-specification#type.
	invalidEventTypeSegment = regexp.MustCompile("[^a-zA-Z0-9.]")
)

const (
	// cleanerName used as the logger name.
	cleanerName = "event-type-cleaner"
)

type Cleaner interface {
	Clean(eventType string) (string, error)
}

type cleaner struct {
	eventTypePrefix   string
	applicationLister *application.Lister // applicationLister will be nil when disabled.
	logger            *logger.Logger
}

// compile-time check.
var _ Cleaner = &cleaner{}

func NewCleaner(eventTypePrefix string, applicationLister *application.Lister, logger *logger.Logger) Cleaner {
	return &cleaner{eventTypePrefix: eventTypePrefix, applicationLister: applicationLister, logger: logger}
}

func (c *cleaner) isApplicationListerEnabled() bool {
	return c.applicationLister != nil
}

// Clean cleans the event-type from none-alphanumeric characters and returns it
// or returns an error if the event-type parsing failed.
func (c *cleaner) Clean(eventType string) (string, error) {
	// format logger
	namedLogger := c.namedLogger(eventType)

	appName, event, version, err := parse(eventType, c.eventTypePrefix)
	if err != nil {
		return "", xerrors.Errorf("failed to parse event-type=%s with prefix=%s: %v", eventType, c.eventTypePrefix, err)
	}

	// clean the application name
	eventTypeClean := build(c.eventTypePrefix, application.GetCleanName(appName), event, version)
	if c.isApplicationListerEnabled() {
		if appObj, err := c.applicationLister.Get(appName); err == nil {
			eventTypeClean = build(c.eventTypePrefix, application.GetCleanTypeOrName(appObj), event, version)
		} else {
			namedLogger.With("application", appName).Debug("Cannot find application")
		}
	}

	// clean the event-type segments
	eventTypeClean = cleanEventType(eventTypeClean)
	namedLogger.With("before", eventType, "after", eventTypeClean).Debug("Clean event-type")

	return eventTypeClean, nil
}

func (c *cleaner) namedLogger(eventType string) *zap.SugaredLogger {
	return c.logger.WithContext().Named(cleanerName).With("prefix", c.eventTypePrefix, "type", eventType)
}

func cleanEventType(eventType string) string {
	return invalidEventTypeSegment.ReplaceAllString(eventType, "")
}
