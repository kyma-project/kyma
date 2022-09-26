package cleaner

import (
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"regexp"
)

var (
	invalidSourceSegment = regexp.MustCompile(`[\s.>*]`)

	invalidEventTypeSegment = regexp.MustCompile(`[\s>*]`)
)

// Perform a compile-time check.
var _ Cleaner = &JetStreamCleaner{}

func NewJetStreamCleaner(logger *logger.Logger) Cleaner {
	return &JetStreamCleaner{logger: logger}
}

func (c *JetStreamCleaner) CleanSource(source string) (string, error) {
	return invalidSourceSegment.ReplaceAllString(source, ""), nil
}

func (c *JetStreamCleaner) CleanEventType(eventType string) (string, error) {
	return invalidEventTypeSegment.ReplaceAllString(eventType, ""), nil
}
