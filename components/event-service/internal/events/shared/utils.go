package shared

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

// ErrorResponseBadRequest returns an error of type PublishEventResponses with BadRequest status code
func ErrorResponseBadRequest(moreInfo string) (response *api.PublishEventResponses) {
	var details []api.ErrorDetail
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeBadPayload, Message: ErrorMessageBadPayload, MoreInfo: moreInfo, Details: details}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}

// ErrorResponseMissingFieldEventType returns an error of type PublishEventResponses for missing EventType field
func ErrorResponseMissingFieldEventType() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldEventType)
}

// ErrorResponseMissingFieldEventTypeVersion returns an error of type PublishEventResponses for missing EventTypeVersion field
func ErrorResponseMissingFieldEventTypeVersion() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldEventTypeVersion)
}

// ErrorResponseWrongEventTypeVersion returns an error of type PublishEventResponses for wrong EventTypeVersion field
func ErrorResponseWrongEventTypeVersion() (response *api.PublishEventResponses) {
	return createInvalidFieldError(FieldEventTypeVersion)
}

// ErrorResponseMissingFieldEventTime returns an error of type PublishEventResponses for missing EventTime field
func ErrorResponseMissingFieldEventTime() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldEventTime)
}

// ErrorResponseWrongEventTime returns an error of type PublishEventResponses for wrong EventTime field
func ErrorResponseWrongEventTime(err error) (response *api.PublishEventResponses) {
	return createInvalidFieldError(FieldEventTime)
}

// ErrorResponseWrongEventID returns an error of type PublishEventResponses for wrong EventID field
func ErrorResponseWrongEventID() (response *api.PublishEventResponses) {
	return createInvalidFieldError(FieldEventID)
}

// ErrorResponseMissingFieldData returns an error of type PublishEventResponses for missing Data field
func ErrorResponseMissingFieldData() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldData)
}

func createMissingFieldError(field interface{}) (response *api.PublishEventResponses) {
	apiErrorDetail := api.ErrorDetail{Field: field.(string), Type: ErrorTypeMissingField, Message: ErrorMessageMissingField, MoreInfo: ""}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeValidationViolation, Message: ErrorMessageMissingField, MoreInfo: "", Details: details}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}

func createInvalidFieldError(field interface{}) (response *api.PublishEventResponses) {
	apiErrorDetail := api.ErrorDetail{Field: field.(string), Type: ErrorTypeInvalidField, Message: ErrorMessageInvalidField, MoreInfo: ""}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeValidationViolation, Message: ErrorMessageInvalidField, MoreInfo: "", Details: details}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}
