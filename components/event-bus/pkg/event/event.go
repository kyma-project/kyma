package event

import (
	"github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/components/event-bus/api/publish"
)

// Further Kyma specific validations in addition to CloudEvents specification
func Validate(event *cloudevents.Event) []publish.ErrorDetail {
	var errors []publish.ErrorDetail
	eventBytes, err := event.DataBytes()
	if err != nil {
		errors = append(errors, publish.ErrorDetail{
			Field:   "data",
			Type:    publish.ErrorTypeBadPayload,
			Message: err.Error(),
		})
	}
	// empty payload is considered as error by earlier /v2 endpoint which was not using cloudevents sdk-go yet
	if len(eventBytes) == 0 {
		errors = append(errors, publish.ErrorDetail{
			Field:   "data",
			Type:    publish.ErrorTypeBadPayload,
			Message: "payload is missing",
		})
	}
	_, err = event.Context.GetExtension(publish.FieldEventTypeVersion)
	if err != nil {
		errors = append(errors, publish.ErrorDetail{
			Field:   publish.FieldEventTypeVersion,
			Type:    publish.ErrorTypeMissingField,
			Message: publish.ErrorMessageMissingField,
		})
	}

	return errors
}
