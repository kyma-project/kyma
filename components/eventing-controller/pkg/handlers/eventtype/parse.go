package eventtype

import (
	"errors"
	"fmt"
	"strings"
)

// parse parses the event-type given the prefix then returns the application name, event and version
// or returns an error if the event-type format is invalid
// a valid even-type format should be: prefix.applicationName.businessObject.operation.version
func parse(eventType, prefix string) (string, string, string, error) {
	if !strings.HasPrefix(eventType, prefix) {
		return "", "", "", errors.New(fmt.Sprintf("failed to parse event-type [%s], prefix not found [%s]", eventType, prefix))
	}

	// remove the event-type prefix
	eventTypeWithoutPrefix := strings.ReplaceAll(eventType, prefix, "")
	eventTypeWithoutPrefix = strings.TrimPrefix(eventTypeWithoutPrefix, ".")

	// make sure that the remaining string has at least 4 segments separated by "." (e.g. myapp.order.created.v1)
	parts := strings.Split(eventTypeWithoutPrefix, ".")
	if len(parts) < 4 {
		return "", "", "", errors.New(fmt.Sprintf("failed to parse event-type [%s], invalid format", eventType))
	}

	// event-type structure is valid, continue parsing
	applicationName := strings.Join(parts[0:len(parts)-3], ".")
	event := strings.Join(parts[len(parts)-3:len(parts)-1], ".")
	version := parts[len(parts)-1]

	return applicationName, event, version, nil
}
