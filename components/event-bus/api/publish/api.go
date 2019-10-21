package publish

const (
	// FieldData value
	FieldData = "data"
	// FieldEventID value
	FieldEventID = "event-id"
	// FieldEventTime value
	FieldEventTime = "event-time"
	// FieldEventType value
	FieldEventType = "event-type"
	// FieldEventTypeVersion value
	FieldEventTypeVersion = "eventtypeversion"
	// FieldSourceID value
	FieldSourceID = "source-id"
	// FieldTraceContext value
	FieldTraceContext = "trace-context"

	// AllowedEventIDChars regex
	AllowedEventIDChars = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`

	// AllowedSourceIDChars regex
	AllowedSourceIDChars = `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$`
	// AllowedEventTypeChars regex
	AllowedEventTypeChars = `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$`
	// AllowedEventTypeVersionChars regex
	AllowedEventTypeVersionChars = `^[a-zA-Z0-9]+$`

	// HeaderSourceID heaver
	HeaderSourceID = "Source-Id"
)

type PublishStatus string

const (
	// PublishFailed status label
	PublishFailed PublishStatus = "failed"
	// PublishIgnored status label
	PublishIgnored PublishStatus = "ignored"
	// PublishPublished status label
	PublishPublished PublishStatus = "published"
)

// Request represents a publish request
type Request struct {
	SourceID           string   `json:"source-id"`
	EventType          string   `json:"event-type"`
	EventTypeVersion   string   `json:"event-type-version"`
	EventID            string   `json:"event-id"`
	EventTime          string   `json:"event-time"`
	Data               AnyValue `json:"data"`
	SourceIDFromHeader bool
}

// AnyValue implements the service definition of AnyValue
type AnyValue interface{}

// Response represents a successful publish response
type Response struct {
	EventID string        `json:"event-id"`
	Status  PublishStatus `json:"status"`
	Reason  string        `json:"reason"`
}

// CloudEvent represents the event to be persisted to NATS
type CloudEvent struct {
	Request
	Extensions Extensions `json:"extensions,omitempty"`
}

// Extensions type
type Extensions = map[string]interface{}

// TraceContext type
type TraceContext map[string]string
