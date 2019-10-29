package trace

const (
	// EventID used for tracing
	EventID          = "event-id"
	// SourceID used for tracing
	SourceID         = "source-id"
	// EventType used for tracing
	EventType        = "event-type"
	// EventTypeVersion used for tracing
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
