package cleaner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

// Perform a compile-time check.
var _ Cleaner = &EventMeshCleaner{}

const (
	minEventMeshSegmentsLimit = 2
	maxEventMeshSegmentsLimit = 3
)

var (
	// invalidEventMeshTypeSegment used to match and replace none-alphanumeric characters in the event-type segments
	// as per SAP Event spec https://github.tools.sap/CentralEngineering/sap-event-specification#type.
	invalidEventMeshTypeSegment   = regexp.MustCompile("[^a-zA-Z0-9.]")
	invalidEventMeshSourceSegment = regexp.MustCompile("[^a-zA-Z0-9]")
)

func NewEventMeshCleaner(logger *logger.Logger) Cleaner {
	return &EventMeshCleaner{logger: logger}
}

func (c *EventMeshCleaner) CleanSource(source string) (string, error) {
	return invalidEventMeshSourceSegment.ReplaceAllString(source, ""), nil
}

func (c *EventMeshCleaner) CleanEventType(eventType string) (string, error) {
	// check for minEventMeshSegmentsLimit
	if len(strings.Split(eventType, ".")) < minEventMeshSegmentsLimit {
		return "", fmt.Errorf("event type should have atlease %d segments", minEventMeshSegmentsLimit)
	}

	mergedEventType := c.getMergedSegments(eventType)
	return invalidEventMeshTypeSegment.ReplaceAllString(mergedEventType, ""), nil
}

// getMergedSegments returns the event type after merging the extra segments
// if the event type contains more than three segments, it combines them into three segments
// (e.g. Account.Root.Created.v1 --> AccountRoot.Created.v1).
func (c *EventMeshCleaner) getMergedSegments(eventType string) string {
	segments := strings.Split(eventType, ".")
	totalSegments := len(segments)
	if totalSegments > maxEventMeshSegmentsLimit {
		combinedSegment := ""
		// combine the first n-2 segments without dots "."
		for i := 0; i < totalSegments-2; i++ { //nolint:gomnd
			combinedSegment += segments[i]
		}
		// append the last  two segment with preceding dots "."
		return fmt.Sprintf("%s.%s.%s", combinedSegment, segments[totalSegments-2], segments[totalSegments-1])
	}

	return eventType
}
