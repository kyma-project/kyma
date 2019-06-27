package v1

import (
	"regexp"
	"time"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
)

var (
	isValidEventID = regexp.MustCompile(api.AllowedEventIDChars).MatchString

	// channel name components
	isValidSourceID         = regexp.MustCompile(api.AllowedSourceIDChars).MatchString
	isValidEventType        = regexp.MustCompile(api.AllowedEventTypeChars).MatchString
	isValidEventTypeVersion = regexp.MustCompile(api.AllowedEventTypeVersionChars).MatchString
)

//ValidatePublish validates a publish POST request
func ValidatePublish(r *api.Request, opts *api.EventOptions) *api.Error {
	if len(r.SourceID) == 0 {
		return api.ErrorResponseMissingFieldSourceID()
	}
	if len(r.EventType) == 0 {
		return api.ErrorResponseMissingFieldEventType()
	}
	if len(r.EventTypeVersion) == 0 {
		return api.ErrorResponseMissingFieldEventTypeVersion()
	}
	if len(r.EventTime) == 0 {
		return api.ErrorResponseMissingFieldEventTime()
	}
	if r.Data == nil {
		return api.ErrorResponseMissingFieldData()
	} else if d, ok := (r.Data).(string); ok && d == "" {
		return api.ErrorResponseMissingFieldData()
	}

	//validate the event components lengths
	if len(r.SourceID) > opts.MaxSourceIDLength {
		return api.ErrorInvalidSourceIDLength(opts.MaxSourceIDLength)
	}
	if len(r.EventType) > opts.MaxEventTypeLength {
		return api.ErrorInvalidEventTypeLength(opts.MaxEventTypeLength)
	}
	if len(r.EventTypeVersion) > opts.MaxEventTypeVersionLength {
		return api.ErrorInvalidEventTypeVersionLength(opts.MaxEventTypeVersionLength)
	}

	// validate the fully-qualified topic name components
	if !isValidSourceID(r.SourceID) {
		return api.ErrorResponseWrongSourceID(r.SourceIDFromHeader)
	}
	if !isValidEventType(r.EventType) {
		return api.ErrorResponseWrongEventType()
	}
	if !isValidEventTypeVersion(r.EventTypeVersion) {
		return api.ErrorResponseWrongEventTypeVersion()
	}

	if _, err := time.Parse(time.RFC3339, r.EventTime); err != nil {
		return api.ErrorResponseWrongEventTime()
	}
	if len(r.EventID) > 0 && !isValidEventID(r.EventID) {
		return api.ErrorResponseWrongEventID()
	}
	return nil
}
