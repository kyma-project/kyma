package publish

const ( //TODO Check how to access struct tags
	FieldData             = "data"
	FieldEventId          = "event-id"
	FieldEventTime        = "event-time"
	FieldEventType        = "event-type"
	FieldEventTypeVersion = "event-type-version"
	FieldSourceId         = "source-id"
	FieldTraceContext     = "trace-context"

	AllowedEventIDChars = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`

	// fully-qualified topic name components
	AllowedSourceIdChars         = `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$`
	AllowedEventTypeChars        = `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$`
	AllowedEventTypeVersionChars = `^[a-zA-Z0-9]+$`

	HeaderSourceId = "Source-Id"
)

// PublishRequest represents a publish request
type PublishRequest struct {
	SourceID           string   `json:"source-id"`
	EventType          string   `json:"event-type"`
	EventTypeVersion   string   `json:"event-type-version"`
	EventID            string   `json:"event-id"`
	EventTime          string   `json:"event-time"`
	Data               AnyValue `json:"data"`
	SourceIdFromHeader bool
}

// AnyValue implements the service definition of AnyValue
type AnyValue interface{}

// PublishResponse represents a successful publish response
type PublishResponse struct {
	EventID string `json:"event-id"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

// CloudEvent represents the event to be persisted to NATS
type CloudEvent struct {
	PublishRequest
	Extensions Extensions `json:"extensions,omitempty"`
}

type Extensions = map[string]interface{}

type TraceContext map[string]string
