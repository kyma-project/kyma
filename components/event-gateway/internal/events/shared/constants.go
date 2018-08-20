package shared

const (
	AllowedEventTypeVersionChars = `[a-zA-Z0-9]+`
	AllowedEventIdChars          = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
)

// Handlers paths
const (
	EventsPath = "/v1/events"
)

// Error messages
const (
	ErrorMessageBadPayload   = "Bad Payload Syntax"
	ErrorMessageMissingField = "Missing Field"
	ErrorMessageInvalidField = "Invalid Field"
)

//Error type definition
const (
	ErrorTypeBadPayload          = "bad_payload_syntax"
	ErrorTypeMissingField        = "missing_field"
	ErrorTypeValidationViolation = "validation_violation"
	ErrorTypeInvalidField        = "invalid_field"
)

// Field definition
const (
	FieldEventId          = "event-id"
	FieldEventTime        = "event-time"
	FieldEventType        = "event-type"
	FieldEventTypeVersion = "event-type-version"
	FieldData             = "data"
)
