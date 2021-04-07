package eventtype

import (
	"errors"
	"fmt"
	"strings"
)

// parse parses the event-type given the prefix then returns the application name, event and version
// or returns an error if the event-type format is invalid
// a valid even-type format should be: prefix.application.event.version
// where event should consist of at least two segments separated by "." (e.g. businessObject.operation)
// constraint: the application segment in the input event-type should not contain "."
func parse(eventType, prefix string) (string, string, string, error) {
	if !strings.HasPrefix(eventType, prefix) {
		return "", "", "", errors.New(fmt.Sprintf("failed to parse event-type [%s], prefix not found [%s]", eventType, prefix))
	}

	// remove the prefix
	eventType = strings.ReplaceAll(eventType, prefix, "")
	eventType = strings.TrimPrefix(eventType, ".")

	// make sure that the remaining string has at least 4 segments separated by "."
	// (e.g. application.businessObject.operation.version)
	parts := strings.Split(eventType, ".")
	if len(parts) < 4 {
		return "", "", "", errors.New(fmt.Sprintf("failed to parse event-type [%s], invalid format", eventType))
	}

	// parse the event-type segments
	applicationName := parts[0]
	businessObject := strings.Join(parts[1:len(parts)-2], "")
	operation := strings.Join(parts[len(parts)-2:len(parts)-1], ".")
	version := parts[len(parts)-1]
	event := fmt.Sprintf("%s.%s", businessObject, operation)

	return applicationName, event, version, nil
}
