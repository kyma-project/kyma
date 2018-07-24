package publish

const ( //TODO Check how to access struct tags
	FieldData              = "data"
	FieldEventId           = "event-id"
	FieldEventTime         = "event-time"
	FieldEventType         = "event-type"
	FieldEventTypeVersion  = "event-type-version"
	FieldSource            = "source"
	FieldSourceType        = "source.source-type"
	FieldSourceNamespace   = "source.source-namespace"
	FieldSourceEnvironment = "source.source-environment"
	FieldTraceContext      = "trace-context"

	//AllowedIDChars
	AllowedIdChars      = `^[a-zA-Z0-9_\-]+$`
	AllowedEventIDChars = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`

	// fully-qualified topic name components
	AllowedSourceEnvironmentChars = `^[a-zA-Z]+([_\-\.]?[a-zA-Z0-9]+)*$`
	AllowedSourceNamespaceChars   = `^[a-zA-Z]+([_\-\.]?[a-zA-Z0-9]+)*$`
	AllowedSourceTypeChars        = `^[a-zA-Z]+([_\-\.]?[a-zA-Z0-9]+)*$`
	AllowedEventTypeChars         = `^[a-zA-Z]+([_\-\.]?[a-zA-Z0-9]+)*$`
	AllowedEventTypeVersionChars  = `^[a-zA-Z0-9]+$`
)

// PublishRequest represents a publish request
type PublishRequest struct {
	Source           *EventSource `json:"source"`
	EventType        string       `json:"event-type"`
	EventTypeVersion string       `json:"event-type-version"`
	EventID          string       `json:"event-id"`
	EventTime        string       `json:"event-time"`
	Data             AnyValue     `json:"data"`
}

// AnyValue implements the service definition of AnyValue
type AnyValue interface{}

// EventSource describes the software instance that emits the event at runtime (i.e. the producer).
type EventSource struct {
	SourceNamespace   string `json:"source-namespace"`
	SourceType        string `json:"source-type"`
	SourceEnvironment string `json:"source-environment"`
}

// PublishResponse represents a successful publish response
type PublishResponse struct {
	EventID string `json:"event-id"`
}

// CloudEvent represents the event to be persisted to NATS
type CloudEvent struct {
	PublishRequest
	Extensions Extensions `json:"extensions,omitempty"`
}

type Extensions = map[string]interface{}

type TraceContext map[string]string
