package shared

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
)

// ErrorResponseBadRequest returns an error of type PublishEventResponse with BadRequest status code
func ErrorResponseBadRequest(moreInfo string) (response *api.PublishEventResponse) {
	var details []api.ErrorDetail
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeBadPayload, Message: ErrorMessageBadPayload, MoreInfo: moreInfo, Details: details}
	return &api.PublishEventResponse{Ok: nil, Error: &apiError}
}

// ErrorResponseRequestBodyTooLarge returns an error of type PublishEventResponse with BadRequest status code
func ErrorResponseRequestBodyTooLarge(moreInfo string) (response *api.PublishEventResponse) {
	var details []api.ErrorDetail
	apiError := api.Error{Status: http.StatusRequestEntityTooLarge, Type: ErrorTypeRequestBodyTooLarge, Message: ErrorMessageRequestBodyTooLarge, MoreInfo: moreInfo, Details: details}
	return &api.PublishEventResponse{Ok: nil, Error: &apiError}
}

// ErrorResponseMissingFieldEventType returns an error of type PublishEventResponse for missing EventType field
func ErrorResponseMissingFieldEventType() (response *api.PublishEventResponse) {
	return CreateMissingFieldError(FieldEventType)
}

// ErrorResponseMissingFieldEventTypeVersion returns an error of type PublishEventResponse for missing EventTypeVersion field
func ErrorResponseMissingFieldEventTypeVersion() (response *api.PublishEventResponse) {
	return CreateMissingFieldError(FieldEventTypeVersion)
}

// ErrorResponseWrongEventTypeVersion returns an error of type PublishEventResponse for wrong EventTypeVersion field
func ErrorResponseWrongEventTypeVersion() (response *api.PublishEventResponse) {
	return CreateInvalidFieldError(FieldEventTypeVersion)
}

// ErrorResponseMissingFieldEventTime returns an error of type PublishEventResponse for missing EventTime field
func ErrorResponseMissingFieldEventTime() (response *api.PublishEventResponse) {
	return CreateMissingFieldError(FieldEventTime)
}

// ErrorResponseWrongEventTime returns an error of type PublishEventResponse for wrong EventTime field
func ErrorResponseWrongEventTime(err error) (response *api.PublishEventResponse) {
	return CreateInvalidFieldError(FieldEventTime)
}

// ErrorResponseWrongEventID returns an error of type PublishEventResponse for wrong EventID field
func ErrorResponseWrongEventID() (response *api.PublishEventResponse) {
	return CreateInvalidFieldError(FieldEventID)
}

// ErrorResponseMissingFieldData returns an error of type PublishEventResponse for missing Data field
func ErrorResponseMissingFieldData() (response *api.PublishEventResponse) {
	return CreateMissingFieldError(FieldData)
}

//CreateMissingFieldError create an error for a missing field
func CreateMissingFieldError(field interface{}) (response *api.PublishEventResponse) {
	apiErrorDetail := api.ErrorDetail{Field: field.(string), Type: ErrorTypeMissingField, Message: ErrorMessageMissingField, MoreInfo: ""}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeValidationViolation, Message: ErrorMessageMissingField, MoreInfo: "", Details: details}
	return &api.PublishEventResponse{Ok: nil, Error: &apiError}
}

//CreateInvalidFieldError creates an error for an invalid field
func CreateInvalidFieldError(field interface{}) (response *api.PublishEventResponse) {
	apiErrorDetail := api.ErrorDetail{Field: field.(string), Type: ErrorTypeInvalidField, Message: ErrorMessageInvalidField, MoreInfo: ""}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeValidationViolation, Message: ErrorMessageInvalidField, MoreInfo: "", Details: details}
	return &api.PublishEventResponse{Ok: nil, Error: &apiError}
}
