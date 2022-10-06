package cleaner

import (
	"regexp"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

var (
	invalidSourceCharacters = regexp.MustCompile(`[\s.>*]`)

	invalidEventTypeCharacters = regexp.MustCompile(`[\s>*]`)
)

// Perform a compile-time check.
var _ Cleaner = &JetStreamCleaner{}

func NewJetStreamCleaner(logger *logger.Logger) Cleaner {
	return &JetStreamCleaner{logger: logger}
}

func (c *JetStreamCleaner) CleanSource(source string) (string, error) {
	return invalidSourceCharacters.ReplaceAllString(source, ""), nil
}

func (c *JetStreamCleaner) CleanEventType(eventType string) (string, error) {
	return invalidEventTypeCharacters.ReplaceAllString(eventType, ""), nil
}
