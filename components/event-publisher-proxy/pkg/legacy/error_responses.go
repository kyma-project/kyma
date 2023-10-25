package legacy

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
)

// An HTTPErrorResponse represents an error with a status code and an error message.
type HTTPErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

// ErrorResponseBadRequest returns an error of type PublishEventResponses with BadRequest status code.
func ErrorResponseBadRequest(moreInfo string) *api.PublishEventResponses {
	var details []api.ErrorDetail
	apiError := api.Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeBadPayload,
		Message:  ErrorMessageBadPayload,
		MoreInfo: moreInfo,
		Details:  details,
	}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}

// ErrorResponseRequestBodyTooLarge returns an error of type PublishEventResponses with BadRequest status code.
func ErrorResponseRequestBodyTooLarge(moreInfo string) *api.PublishEventResponses {
	var details []api.ErrorDetail
	apiError := api.Error{
		Status:   http.StatusRequestEntityTooLarge,
		Type:     ErrorTypeRequestBodyTooLarge,
		Message:  ErrorMessageRequestBodyTooLarge,
		MoreInfo: moreInfo,
		Details:  details,
	}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}

// ErrorResponseMissingFieldEventType returns an error of type PublishEventResponses for missing EventType field.
func ErrorResponseMissingFieldEventType() *api.PublishEventResponses {
	return CreateMissingFieldError(FieldEventType)
}

// ErrorResponseMissingFieldEventTypeVersion returns an error of type PublishEventResponses
// for missing EventTypeVersion field.
func ErrorResponseMissingFieldEventTypeVersion() *api.PublishEventResponses {
	return CreateMissingFieldError(FieldEventTypeVersion)
}

// ErrorResponseWrongEventTypeVersion returns an error of type PublishEventResponses for wrong EventTypeVersion field.
func ErrorResponseWrongEventTypeVersion() *api.PublishEventResponses {
	return CreateInvalidFieldError(FieldEventTypeVersion)
}

// ErrorResponseMissingFieldEventTime returns an error of type PublishEventResponses for missing EventTime field.
func ErrorResponseMissingFieldEventTime() *api.PublishEventResponses {
	return CreateMissingFieldError(FieldEventTime)
}

// ErrorResponseWrongEventTime returns an error of type PublishEventResponses for wrong EventTime field.
func ErrorResponseWrongEventTime() *api.PublishEventResponses {
	return CreateInvalidFieldError(FieldEventTime)
}

// ErrorResponseWrongEventID returns an error of type PublishEventResponses for wrong EventID field.
func ErrorResponseWrongEventID() *api.PublishEventResponses {
	return CreateInvalidFieldError(FieldEventID)
}

// ErrorResponseMissingFieldData returns an error of type PublishEventResponses for missing Data field.
func ErrorResponseMissingFieldData() *api.PublishEventResponses {
	return CreateMissingFieldError(FieldData)
}

// ErrorResponse returns an error of type PublishEventResponses with the given status and error.
func ErrorResponse(status int, err error) *api.PublishEventResponses {
	return &api.PublishEventResponses{Error: &api.Error{Status: status, Message: err.Error()}}
}

// CreateMissingFieldError creates an error for a missing field.
func CreateMissingFieldError(field any) *api.PublishEventResponses {
	apiErrorDetail := api.ErrorDetail{
		Field:    field.(string),
		Type:     ErrorTypeMissingField,
		Message:  ErrorMessageMissingField,
		MoreInfo: "",
	}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeValidationViolation,
		Message:  ErrorMessageMissingField,
		MoreInfo: "",
		Details:  details,
	}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}

// CreateInvalidFieldError creates an error for an invalid field.
func CreateInvalidFieldError(field any) *api.PublishEventResponses {
	apiErrorDetail := api.ErrorDetail{
		Field:    field.(string),
		Type:     ErrorTypeInvalidField,
		Message:  ErrorMessageInvalidField,
		MoreInfo: "",
	}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeValidationViolation,
		Message:  ErrorMessageInvalidField,
		MoreInfo: "",
		Details:  details,
	}
	return &api.PublishEventResponses{Ok: nil, Error: &apiError}
}
