package v2

const (
	// FieldData value
	FieldData = "data"
	// FieldEventID value
	FieldEventID = "id"
	// FieldEventTime value
	FieldEventTime = "time"
	// FieldEventType value
	FieldEventType = "type"
	// FieldSpecVersion value
	FieldSpecVersion = "specversion"
	// FieldEventTypeVersion value
	FieldEventTypeVersion = "eventtypeversion"
	// FieldSourceID value
	FieldSourceID = "source"

	// AllowedEventIDChars regex
	AllowedEventIDChars = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`

	// AllowedEventTypeVersionChars regex
	AllowedEventTypeVersionChars = `^[a-zA-Z0-9]+$`

	//SpecVersionV3 Value
	SpecVersionV3 = "0.3"
)

// Extensions type
type Extensions = map[string]interface{}

// AnyValue implements the service definition of AnyValue
type AnyValue interface{}

// EventRequestV2 represents a publish event CE v.3.0 request
type EventRequestV2 struct {
	ID                  string   `json:"id"`
	Source              string   `json:"source"`
	SpecVersion         string   `json:"specversion"`
	Type                string   `json:"type"`
	DataContentEncoding string   `json:"datacontentencoding,omitempty"`
	TypeVersion         string   `json:"eventtypeversion"`
	Time                string   `json:"time"`
	Data                AnyValue `json:"data"`
}

// CloudEventV3 represents the event to be persisted to NATS
type CloudEventV3 struct {
	EventRequestV2
	Extensions Extensions `json:"extensions,omitempty"`
}
