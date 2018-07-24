package shared

import (
	"net/http"

	"github.com/kyma-project/kyma/components/gateway/internal/events/api"
)

func ErrorResponseBadRequest(moreInfo string) (response *api.PublishEventResponses) {
	details := []api.ErrorDetail{}
	apiError := api.Error{Status: http.StatusBadRequest, Type: ErrorTypeBadPayload, Message: ErrorMessageBadPayload, MoreInfo: moreInfo, Details: details}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}

func ErrorResponseMissingFieldEventType() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldEventType)
}

func ErrorResponseMissingFieldEventTypeVersion() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldEventTypeVersion)
}

func ErrorResponseWrongEventTypeVersion() (response *api.PublishEventResponses) {
	return createInvalidFieldError(FieldEventTypeVersion)
}

func ErrorResponseMissingFieldEventTime() (response *api.PublishEventResponses) {
	return createMissingFieldError(FieldEventTime)
}

func ErrorResponseWrongEventTime(err error) (response *api.PublishEventResponses) {
	return createInvalidFieldError(FieldEventTime)
}

func ErrorResponseWrongEventId() (response *api.PublishEventResponses) {
	return createInvalidFieldError(FieldEventId)
}

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
