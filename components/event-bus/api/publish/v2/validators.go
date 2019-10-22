package v2

import (
	"github.com/cloudevents/sdk-go"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"regexp"
)

var (
	// channel name components
	// TODO(nachtmaar): only used by tests
	isValidSourceID         = regexp.MustCompile(api.AllowedSourceIDChars).MatchString
	isValidEventType        = regexp.MustCompile(api.AllowedEventTypeChars).MatchString
	isValidEventTypeVersion = regexp.MustCompile(api.AllowedEventTypeVersionChars).MatchString
)

// Further Kyma specific validations in addition to CloudEvents specification
func ValidateKyma(event *cloudevents.Event) []api.ErrorDetail {
	var errors []api.ErrorDetail
	eventBytes, err := event.DataBytes()
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   "data",
			Type:    api.ErrorTypeBadPayload,
			Message: err.Error(),
		})
	}
	// empty payload is considered as error by earlier /v2 endpoint which was not using cloudevents sdk-go yet
	if len(eventBytes) == 0 {
		errors = append(errors, api.ErrorDetail{
			Field:   "data",
			Type:    api.ErrorTypeBadPayload,
			Message: "payload is missing",
		})
	}
	_, err = event.Context.GetExtension(api.FieldEventTypeVersion)
	if err != nil {
		errors = append(errors, api.ErrorDetail{
			Field:   api.FieldEventTypeVersion,
			Type:    api.ErrorTypeMissingField,
			Message: api.ErrorMessageMissingField,
		})
	}

	return errors
}
