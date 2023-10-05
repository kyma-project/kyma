package builder

import (
	"fmt"
	"strings"

	cev2event "github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
)

// Perform a compile-time check.
var _ CloudEventBuilder = &GenericBuilder{}

var (
	// jsBuilderName used as the logger name.
	genericBuilderName = "generic-type-builder"
)

func NewGenericBuilder(typePrefix string, cleaner cleaner.Cleaner, applicationLister *application.Lister, logger *logger.Logger) CloudEventBuilder {
	return &GenericBuilder{
		typePrefix:        typePrefix,
		applicationLister: applicationLister,
		logger:            logger,
		cleaner:           cleaner,
	}
}

func (gb *GenericBuilder) isApplicationListerEnabled() bool {
	return gb.applicationLister != nil
}

func (gb *GenericBuilder) Build(event cev2event.Event) (*cev2event.Event, error) {
	// format logger
	namedLogger := gb.namedLogger(event.Source(), event.Type())

	// clean the source
	cleanSource, err := gb.cleaner.CleanSource(gb.GetAppNameOrSource(event.Source(), namedLogger))
	if err != nil {
		return nil, err
	}

	// clean the event type
	cleanEventType, err := gb.cleaner.CleanEventType(event.Type())
	if err != nil {
		return nil, err
	}

	// build event type
	finalEventType := gb.getFinalSubject(cleanSource, cleanEventType)

	// validate if the segments are not empty
	segments := strings.Split(finalEventType, ".")
	if DoesEmptySegmentsExist(segments) {
		return nil, fmt.Errorf("event type cannot have empty segments after cleaning: %s", finalEventType)
	}
	namedLogger.Debugf("using event type: %s", finalEventType)

	ceEvent := event.Clone()
	// set original type header
	ceEvent.SetExtension(OriginalTypeHeaderName, event.Type())
	// set prefixed type
	ceEvent.SetType(finalEventType)
	// validate the final cloud event
	if err = ceEvent.Validate(); err != nil {
		return nil, err
	}

	return &ceEvent, nil
}

// getFinalSubject returns the final prefixed event type.
func (gb *GenericBuilder) getFinalSubject(source, eventType string) string {
	return fmt.Sprintf("%s.%s.%s", gb.typePrefix, source, eventType)
}

// GetAppNameOrSource returns the application name if exists, otherwise returns source name.
func (gb *GenericBuilder) GetAppNameOrSource(source string, namedLogger *zap.SugaredLogger) string {
	var appName = source
	if gb.isApplicationListerEnabled() {
		if appObj, err := gb.applicationLister.Get(source); err == nil && appObj != nil {
			appName = application.GetTypeOrName(appObj)
			namedLogger.With("application", source).Debug("Using application name: %s as source.", appName)
		} else {
			namedLogger.With("application", source).Debug("Cannot find application.")
		}
	}

	return appName
}

func (gb *GenericBuilder) namedLogger(source, eventType string) *zap.SugaredLogger {
	return gb.logger.WithContext().Named(genericBuilderName).With("source", source, "type", eventType)
}
