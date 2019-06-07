package trace

const (
	eventID          = "event-id"
	sourceID         = "source-id"
	eventType        = "event-type"
	eventTypeVersion = "event-type-ver"
)

const (
	// HeaderSourceID header
	HeaderSourceID = "Ce-Source-ID"
	// HeaderEventType header
	HeaderEventType = "Ce-Event-Type"
	// HeaderEventTypeVersion header
	HeaderEventTypeVersion = "Ce-Event-Type-Version"
	// HeaderEventID header
	HeaderEventID = "Ce-Event-ID"
	// HeaderEventTime header
	HeaderEventTime = "Ce-Event-Time"
)

// CreateTraceTagsFromMessageHeader returns a map of trace headers.
func CreateTraceTagsFromMessageHeader(headers map[string][]string) map[string]string {
	return map[string]string{
		eventID:          headers[HeaderEventID][0],
		sourceID:         headers[HeaderSourceID][0],
		eventType:        headers[HeaderEventType][0],
		eventTypeVersion: headers[HeaderEventTypeVersion][0],
	}
}
