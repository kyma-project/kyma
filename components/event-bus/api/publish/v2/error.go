package v2

import (
	"net/http"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
)

const (
	// ErrorMessageMissingSourceID represents the error message for `ErrorTypeMissingFieldOrHeader`
	ErrorMessageMissingSourceID = "missing 'source' field in the json payload"
)

// ErrorResponseMissingFieldSpecVersion returns an API error instance for the missing field spec version error.
func ErrorResponseMissingFieldSpecVersion() (response *api.Error) {
	return api.CreateMissingFieldError(FieldSpecVersion)
}

// ErrorResponseWrongSpecVersion returns an API error instance for the wrong spec version error.
func ErrorResponseWrongSpecVersion() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldSpecVersion)
}

// ErrorResponseWrongEventTime returns an API error instance for the wrong event time error.
func ErrorResponseWrongEventTime() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldEventTime)
}

// ErrorResponseMissingFieldSourceID returns an API error instance for the missing field source ID error.
func ErrorResponseMissingFieldSourceID() (response *api.Error) {
	apiErrorDetail := api.ErrorDetail{
		Field:    FieldSourceID,
		Type:     api.ErrorTypeMissingField,
		Message:  ErrorMessageMissingSourceID,
		MoreInfo: "",
	}
	details := []api.ErrorDetail{apiErrorDetail}
	apiError := api.Error{Status: http.StatusBadRequest, Type: api.ErrorTypeValidationViolation,
		Message: ErrorMessageMissingSourceID, MoreInfo: "", Details: details}

	return &apiError
}

// ErrorResponseMissingFieldEventID returns an API error instance for the missing field event type error.
func ErrorResponseMissingFieldEventID() (response *api.Error) {
	return api.CreateMissingFieldError(FieldEventID)
}

// ErrorResponseMissingFieldEventType returns an API error instance for the missing field event type error.
func ErrorResponseMissingFieldEventType() (response *api.Error) {
	return api.CreateMissingFieldError(FieldEventType)
}

// ErrorResponseMissingFieldEventTypeVersion returns an API error instance for the missing field event type version error.
func ErrorResponseMissingFieldEventTypeVersion() (response *api.Error) {
	return api.CreateMissingFieldError(FieldEventTypeVersion)
}

// ErrorResponseMissingFieldEventTime returns an API error instance for the missing field event time error.
func ErrorResponseMissingFieldEventTime() (response *api.Error) {
	return api.CreateMissingFieldError(FieldEventTime)
}

// ErrorResponseWrongEventType returns an API error instance for the wrong event type error.
func ErrorResponseWrongEventType() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldEventType)
}

// ErrorResponseWrongEventTypeVersion returns an API error instance for the wrong event type version error.
func ErrorResponseWrongEventTypeVersion() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldEventTypeVersion)
}

// ErrorResponseWrongEventID returns an API error instance for the wrong event ID error.
func ErrorResponseWrongEventID() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldEventID)
}

// ErrorResponseWrongSourceID returns an API error instance for the wrong source ID error.
func ErrorResponseWrongSourceID() (response *api.Error) {
	return api.CreateInvalidFieldError(FieldSourceID)
}
