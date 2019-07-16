package publish

import (
	"fmt"
	"net/http"
)

const (
	/*ErrorTypeBadPayload The request payload has incorrect syntax according to the sent Content-Type.
	Check the payload content for syntax errors, such as missing commas or quotation marks that are not closed.
	*/
	ErrorTypeBadPayload = "bad_payload_syntax"
	/*ErrorTypeValidationViolation Top level validation error.
	 */
	ErrorTypeValidationViolation = "validation_violation"
	/*ErrorTypeMissingField Sub-level error type of `ErrorTypeValidationViolation` representaing that the requested body
	payload for a POST or PUT operation is missing,	which violates the defined validation constraints. This denotes
	a missing field when a value is expected.
	*/
	ErrorTypeMissingField = "missing_field"

	// ErrorTypeMissingFieldOrHeader error type for a missing field or header.
	ErrorTypeMissingFieldOrHeader = "missing_field/missing_header"
	/*ErrorTypeInvalidField Sub-level error type of `ErrorTypeValidationViolation` representaing that the requested body
	payload for the POST or PUT operation violates the validation constraints.
	This denotes specifically that there is:
	- A type incompatibility, such as a field modeled to be an integer, but a non-numeric expression was found instead.
	- A range under or over flow validation violation cause.
	*/
	ErrorTypeInvalidField = "invalid_field"
	// ErrorTypeInvalidHeader is error type code for invalid header
	ErrorTypeInvalidHeader = "invalid_header"
	// ErrorTypeInvalidFieldLength is error type code for invalid field length
	ErrorTypeInvalidFieldLength = "invalid_field_length"
	// ErrorTypeRequestBodyTooLarge is error type code for error responses where request body is too large
	ErrorTypeRequestBodyTooLarge = "request_body_too_large"
	// ErrorMessageRequestBodyTooLarge is error message for error responses where request body is too large
	ErrorMessageRequestBodyTooLarge = "Request body too large"
	// ErrorTypeInternalServerError Some unexpected internal error occurred while processing the request.
	ErrorTypeInternalServerError = "internal_server_error"
	// ErrorMessageInternalServerError represents the error message for `ErrorTypeInternalServerError`
	ErrorMessageInternalServerError = "Some unexpected internal error occurred, please contact support."
	/*ErrorTypeBadRequest A generic error for bad requests sent by the clients. Use when none of the specific
	error types apply.
	*/
	ErrorTypeBadRequest = "bad_request"
	// ErrorMessageBadRequest represents the error message for `ErrorTypeBadRequest`
	ErrorMessageBadRequest = "Some unexpected internal error occurred, please contact support."
	// ErrorMessageBadPayload represents the error message for `ErrorTypeBadPayload`
	ErrorMessageBadPayload = "Something went very wrong. Please try again."
	// ErrorMessageMissingField represents the error message for `ErrorTypeMissingField`
	ErrorMessageMissingField = "We need all required fields complete to keep you moving."
	// ErrorMessageInvalidField represents the error message for `ErrorTypeInvalidField`
	ErrorMessageInvalidField = "We need all your entries to be correct to keep you moving."
	// ErrorMessageInvalidFieldLength represents the error message for `ErrorTypeInvalidFieldLength`
	ErrorMessageInvalidFieldLength = "Field length must be at max: %d"
	// ErrorMessageMissingSourceID represents the error message for `ErrorTypeMissingFieldOrHeader`
	ErrorMessageMissingSourceID = "Either provide 'Source-Id' header or specify 'source-id' in the json payload"
)

// ErrorDetail represents error cause
type ErrorDetail struct {
	Field    string `json:"field"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	MoreInfo string `json:"moreInfo"`
}

// Error represents API error response code
type Error struct {
	Status   int           `json:"status"`
	Type     string        `json:"type"`
	Message  string        `json:"message"`
	MoreInfo string        `json:"moreInfo"`
	Details  []ErrorDetail `json:"details"`
}

// ErrorResponseInternalServer creates API Error response for case of internal server error.
func ErrorResponseInternalServer() (response *Error) {
	apiError := Error{
		Status:   http.StatusInternalServerError,
		Type:     ErrorTypeInternalServerError,
		Message:  ErrorMessageInternalServerError,
		MoreInfo: "",
	}
	return &apiError
}

// ErrorResponseRequestBodyTooLarge creates API Error response for case of request body being too large
func ErrorResponseRequestBodyTooLarge() (response *Error) {
	apiError := &Error{
		Status:   http.StatusRequestEntityTooLarge,
		Type:     ErrorTypeRequestBodyTooLarge,
		Message:  ErrorMessageRequestBodyTooLarge,
		MoreInfo: "",
	}
	return apiError
}

// ErrorInvalidSourceIDLength creates an API Error response in case Source ID length exceeded the maximum
func ErrorInvalidSourceIDLength(sourceIDMaxLength int) *Error {
	return ErrorInvalidFieldLength(FieldSourceID, sourceIDMaxLength)
}

// ErrorInvalidEventTypeLength creates an API Error response in case Event Type length exceeded the maximum
func ErrorInvalidEventTypeLength(eventTypeMaxLength int) *Error {
	return ErrorInvalidFieldLength(FieldEventType, eventTypeMaxLength)
}

// ErrorInvalidEventTypeVersionLength creates an API Error response in case Event Type Version length exceeded the maximum
func ErrorInvalidEventTypeVersionLength(eventTypeVersionMaxLength int) *Error {
	return ErrorInvalidFieldLength(FieldEventTypeVersion, eventTypeVersionMaxLength)
}

// ErrorInvalidFieldLength returns an API error instance for the invalid field length error.
func ErrorInvalidFieldLength(field string, length int) *Error {
	apiErrorDetail := ErrorDetail{
		Field:    field,
		Type:     ErrorTypeInvalidFieldLength,
		Message:  ErrorMessageInvalidField,
		MoreInfo: fmt.Sprintf(ErrorMessageInvalidFieldLength, length),
	}
	details := []ErrorDetail{apiErrorDetail}
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeValidationViolation,
		Message:  ErrorMessageInvalidField,
		MoreInfo: "",
		Details:  details,
	}
	return &apiError
}

// ErrorResponseBadRequest returns an API error instance for the bad request error.
func ErrorResponseBadRequest() (response *Error) {
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeBadRequest,
		Message:  ErrorMessageBadRequest,
		MoreInfo: "",
	}
	return &apiError
}

// ErrorResponseBadPayload returns an API error instance for the bad payload error.
func ErrorResponseBadPayload() (response *Error) {
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeBadPayload,
		Message:  ErrorMessageBadRequest,
		MoreInfo: "",
	}
	return &apiError
}

// ErrorResponseMissingFieldSourceID returns an API error instance for the missing field source ID error.
func ErrorResponseMissingFieldSourceID() (response *Error) {
	apiErrorDetail := ErrorDetail{
		Field:    FieldSourceID + "/" + HeaderSourceID,
		Type:     ErrorTypeMissingFieldOrHeader,
		Message:  ErrorMessageMissingSourceID,
		MoreInfo: "",
	}
	details := []ErrorDetail{apiErrorDetail}
	apiError := Error{Status: http.StatusBadRequest, Type: ErrorTypeValidationViolation, Message: ErrorMessageMissingSourceID, MoreInfo: "", Details: details}

	return &apiError
}

// ErrorResponseMissingFieldData returns an API error instance for the missing field data error.
func ErrorResponseMissingFieldData() (response *Error) {
	return CreateMissingFieldError(FieldData)
}

// ErrorResponseMissingFieldEventType returns an API error instance for the missing field event type error.
func ErrorResponseMissingFieldEventType() (response *Error) {
	return CreateMissingFieldError(FieldEventType)
}

// ErrorResponseMissingFieldEventTypeVersion returns an API error instance for the missing field event type version error.
func ErrorResponseMissingFieldEventTypeVersion() (response *Error) {
	return CreateMissingFieldError(FieldEventTypeVersion)
}

// ErrorResponseMissingFieldEventTime returns an API error instance for the missing field event time error.
func ErrorResponseMissingFieldEventTime() (response *Error) {
	return CreateMissingFieldError(FieldEventTime)
}

// ErrorResponseWrongEventType returns an API error instance for the wrong event type error.
func ErrorResponseWrongEventType() (response *Error) {
	return CreateInvalidFieldError(FieldEventType)
}

// ErrorResponseWrongEventTypeVersion returns an API error instance for the wrong event type version error.
func ErrorResponseWrongEventTypeVersion() (response *Error) {
	return CreateInvalidFieldError(FieldEventTypeVersion)
}

// ErrorResponseWrongEventTime returns an API error instance for the wrong event time error.
func ErrorResponseWrongEventTime() (response *Error) {
	return CreateInvalidFieldError(FieldEventTime)
}

// ErrorResponseWrongEventID returns an API error instance for the wrong event ID error.
func ErrorResponseWrongEventID() (response *Error) {
	return CreateInvalidFieldError(FieldEventID)
}

// ErrorResponseWrongSourceID returns an API error instance for the wrong source ID error.
func ErrorResponseWrongSourceID(sourceIDFromHeader bool) (response *Error) {
	if sourceIDFromHeader {
		return createInvalidFieldErrorWithType(HeaderSourceID, ErrorTypeInvalidHeader)
	}
	return CreateInvalidFieldError(FieldSourceID)
}

// CreateMissingFieldError creates an Error for a missing field
func CreateMissingFieldError(field interface{}) (response *Error) {
	apiErrorDetail := ErrorDetail{
		Field:    field.(string),
		Type:     ErrorTypeMissingField,
		Message:  ErrorMessageMissingField,
		MoreInfo: "",
	}
	details := []ErrorDetail{apiErrorDetail}
	apiError := Error{Status: http.StatusBadRequest, Type: ErrorTypeValidationViolation, Message: ErrorMessageMissingField, MoreInfo: "", Details: details}
	return &apiError
}

// CreateInvalidFieldError creates an Error for an invalid field
func CreateInvalidFieldError(field interface{}) (response *Error) {
	return createInvalidFieldErrorWithType(field, ErrorTypeInvalidField)
}

func createInvalidFieldErrorWithType(field interface{}, errorType string) (response *Error) {
	apiErrorDetail := ErrorDetail{
		Field:    field.(string),
		Type:     errorType,
		Message:  ErrorMessageInvalidField,
		MoreInfo: "",
	}
	details := []ErrorDetail{apiErrorDetail}
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeValidationViolation,
		Message:  ErrorMessageInvalidField,
		MoreInfo: "",
		Details:  details,
	}
	return &apiError
}
