package publish

import (
	"regexp"
	"time"
)

var (
	isValidID1     = regexp.MustCompile(AllowedIdChars).MatchString
	isValidEventID = regexp.MustCompile(AllowedEventIDChars).MatchString

	// fully-qualified topic name components
	isValidSourceEnvironment = regexp.MustCompile(AllowedSourceEnvironmentChars).MatchString
	isValidSourceNamespace   = regexp.MustCompile(AllowedSourceNamespaceChars).MatchString
	isValidSourceType        = regexp.MustCompile(AllowedSourceTypeChars).MatchString
	isValidEventType         = regexp.MustCompile(AllowedEventTypeChars).MatchString
	isValidEventTypeVersion  = regexp.MustCompile(AllowedEventTypeVersionChars).MatchString
)

//ValidatePublish validates a publish POST request
func ValidatePublish(r *PublishRequest) *Error {
	if r.Source == nil {
		return ErrorResponseMissingFieldSource()
	}
	if r.Source.SourceType == "" {
		return ErrorResponseMissingFieldSourceType()
	}
	if r.Source.SourceNamespace == "" {
		return ErrorResponseMissingFieldSourceNamespace()
	}
	if r.Source.SourceEnvironment == "" {
		return ErrorResponseMissingFieldSourceEnvironment()
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

	// validate the fully-qualified topic name components
	if !isValidSourceEnvironment(r.Source.SourceEnvironment) {
		return ErrorResponseWrongSourceEnvironment()
	}
	if !isValidSourceNamespace(r.Source.SourceNamespace) {
		return ErrorResponseWrongSourceNamespace()
	}
	if !isValidSourceType(r.Source.SourceType) {
		return ErrorResponseWrongSourceType()
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
