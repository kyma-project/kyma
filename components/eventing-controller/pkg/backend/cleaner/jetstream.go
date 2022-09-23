package cleaner

import "github.com/kyma-project/kyma/components/eventing-controller/logger"

// Perform a compile-time check.
var _ Cleaner = &JetStreamCleaner{}

func NewJetStreamCleaner(logger *logger.Logger) Cleaner {
	return &JetStreamCleaner{logger: logger}
}

func (c *JetStreamCleaner) CleanSource(source string) (string, error) {
	return source, nil
}

func (c *JetStreamCleaner) CleanEventType(eventType string) (string, error) {
	return eventType, nil
}
