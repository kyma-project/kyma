package eventtype

import (
	"errors"
	"fmt"
	"strings"
)

// parse splits the event-type using the given prefix and returns the application name, event and version
// or an error if the event-type format is invalid.
// A valid even-type format should be: prefix.application.event.version or application.event.version
// where event should consist of at least two segments separated by "." (e.g. businessObject.operation).
// Constraint: the application segment in the input event-type should not contain ".".
func parse(eventType, prefix string) (string, string, string, error) {
	if !strings.HasPrefix(eventType, prefix) {
		return "", "", "", errors.New("prefix not found")
	}

	// remove the prefix
	eventType = strings.ReplaceAll(eventType, prefix, "")
	eventType = strings.TrimPrefix(eventType, ".")

	// make sure that the remaining string has at least 4 segments separated by "."
	// (e.g. application.businessObject.operation.version)
	parts := strings.Split(eventType, ".")
	if len(parts) < 4 {
		return "", "", "", errors.New("invalid format")
	}

	// parse the event-type segments
	applicationName := parts[0]
	businessObject := strings.Join(parts[1:len(parts)-2], "") // combine segments
	operation := parts[len(parts)-2]
	version := parts[len(parts)-1]
	event := fmt.Sprintf("%s.%s", businessObject, operation)

	return applicationName, event, version, nil
}
