package trace

const (
	eventID          = "event-id"
	sourceID         = "source-id"
	eventType        = "event-type"
	eventTypeVersion = "event-type-ver"
)

// SpecVersion CE header value
const SpecVersion = "0.3"

const (
	// HeaderSourceID header
	HeaderSourceID = "ce-source"
	// HeaderEventType header
	HeaderEventType = "ce-type"
	// HeaderEventTypeVersion header
	HeaderEventTypeVersion = "ce-eventtypeversion"
	// HeaderEventID header
	HeaderEventID = "ce-id"
	// HeaderEventTime header
	HeaderEventTime = "ce-time"
	// HeaderSpecVersion header
	HeaderSpecVersion = "ce-specversion"
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
