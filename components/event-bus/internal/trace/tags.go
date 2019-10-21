package trace

const (
	EventID          = "event-id"
	SourceID         = "source-id"
	EventType        = "event-type"
	EventTypeVersion = "event-type-ver"
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
		EventID:          headers[HeaderEventID][0],
		SourceID:         headers[HeaderSourceID][0],
		EventType:        headers[HeaderEventType][0],
		EventTypeVersion: headers[HeaderEventTypeVersion][0],
	}
}
