package cleaner

import "github.com/kyma-project/kyma/components/eventing-controller/logger"

// Perform a compile-time check.
var _ Cleaner = &EventMeshCleaner{}

func NewEventMeshCleaner(logger *logger.Logger) Cleaner {
	return &EventMeshCleaner{logger: logger}
}

func (c *EventMeshCleaner) CleanSource(source string) (string, error) {
	return source, nil
}

func (c *EventMeshCleaner) CleanEventType(eventType string) (string, error) {
	return eventType, nil
}
