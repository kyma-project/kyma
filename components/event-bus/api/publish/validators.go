package publish

import (
	"regexp"
	"time"
)

const (
	// event components max lengths
	// todo make them all configurable
	SourceIdMaxLength         = 28
	EventTypeMaxLength        = 28
	eventTypeVersionMaxLength = 2
)

var (
	isValidEventID = regexp.MustCompile(AllowedEventIDChars).MatchString

	// fully-qualified topic name components
	isValidSourceId         = regexp.MustCompile(AllowedSourceIdChars).MatchString
	isValidEventType        = regexp.MustCompile(AllowedEventTypeChars).MatchString
	isValidEventTypeVersion = regexp.MustCompile(AllowedEventTypeVersionChars).MatchString
)

//ValidatePublish validates a publish POST request
func ValidatePublish(r *PublishRequest) *Error {
	if len(r.SourceID) == 0 {
		return ErrorResponseMissingFieldSourceId()
	}
	if len(r.EventType) == 0 {
		return ErrorResponseMissingFieldEventType()
	}
	if len(r.EventTypeVersion) == 0 {
		return ErrorResponseMissingFieldEventTypeVersion()
	}
	if len(r.EventTime) == 0 {
		return ErrorResponseMissingFieldEventTime()
	}
	if r.Data == nil {
		return ErrorResponseMissingFieldData()
	} else if d, ok := (r.Data).(string); ok && d == "" {
		return ErrorResponseMissingFieldData()
	}

	//validate the event components lengths
	if len(r.SourceID) > SourceIdMaxLength {
		return errorInvalidSourceIDLength(SourceIdMaxLength)
	}
	if len(r.EventType) > EventTypeMaxLength {
		return errorInvalidEventTypeLength(EventTypeMaxLength)
	}
	if len(r.EventTypeVersion) > eventTypeVersionMaxLength {
		return errorInvalidEventTypeVersionLength(eventTypeVersionMaxLength)
	}

	// validate the fully-qualified topic name components
	if !isValidSourceId(r.SourceID) {
		return ErrorResponseWrongSourceId(r.SourceIdFromHeader)
	}
	if !isValidEventType(r.EventType) {
		return ErrorResponseWrongEventType()
	}
	if !isValidEventTypeVersion(r.EventTypeVersion) {
		return ErrorResponseWrongEventTypeVersion()
	}

	if _, err := time.Parse(time.RFC3339, r.EventTime); err != nil {
		return ErrorResponseWrongEventTime(err)
	}
	if len(r.EventID) > 0 && !isValidEventID(r.EventID) {
		return ErrorResponseWrongEventId()
	}
	return nil
}
