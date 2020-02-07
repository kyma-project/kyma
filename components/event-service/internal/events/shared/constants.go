package shared

// Allowed patterns for the Event components
const (
	AllowedEventTypeVersionChars = `[a-zA-Z0-9]+`
	AllowedEventIDChars          = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
)

// Handlers paths
const (
	EventsV2Path = "/v2/events"
)

// Error messages
const (
	ErrorMessageBadPayload          = "Bad payload syntax"
	ErrorMessageRequestBodyTooLarge = "Request body too large"
	ErrorMessageMissingField        = "Missing field"
	ErrorMessageInvalidField        = "Invalid field"
)

// Error type definition
const (
	ErrorTypeBadPayload          = "bad_payload_syntax"
	ErrorTypeRequestBodyTooLarge = "request_body_too_large"
	ErrorTypeMissingField        = "missing_field"
	ErrorTypeValidationViolation = "validation_violation"
	ErrorTypeInvalidField        = "invalid_field"
)

// Field definition
const (
	FieldEventID          = "event-id"
	FieldEventTime        = "event-time"
	FieldEventType        = "event-type"
	FieldEventTypeVersion = "event-type-version"

	FieldEventIDV2          = "id"
	FieldEventTimeV2        = "time"
	FieldEventTypeV2        = "type"
	FieldEventTypeVersionV2 = "eventtypeversion"
	FieldSpecVersionV2      = "specversion"

	FieldData = "data"
)
