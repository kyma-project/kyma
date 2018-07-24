package publish

import "net/http"

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
	/*ErrorTypeInvalidField Sub-level error type of `ErrorTypeValidationViolation` representaing that the requested body
	payload for the POST or PUT operation violates the validation constraints.
	This denotes specifically that there is:
	- A type incompatibility, such as a field modeled to be an integer, but a non-numeric expression was found instead.
	- A range under or over flow validation violation cause.
	*/
	ErrorTypeInvalidField = "invalid_field"
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

// TODO Add propper comments
func ErrorResponseInternalServer() (response *Error) {
	apiError := Error{
		Status:   http.StatusInternalServerError,
		Type:     ErrorTypeInternalServerError,
		Message:  ErrorMessageInternalServerError,
		MoreInfo: "",
	}
	return &apiError
}

func ErrorResponseBadRequest() (response *Error) {
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeBadRequest,
		Message:  ErrorMessageBadRequest,
		MoreInfo: "",
	}
	return &apiError
}

func ErrorResponseBadPayload() (response *Error) {
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeBadPayload,
		Message:  ErrorMessageBadRequest,
		MoreInfo: "",
	}
	return &apiError
}

func ErrorResponseEmptyRequest() (response *Error) {
	apiErrorDetail := ErrorDetail{
		Field:    "",
		Type:     ErrorTypeInvalidField,
		Message:  ErrorMessageInvalidField,
		MoreInfo: "",
	}
	details := []ErrorDetail{apiErrorDetail}
	apiError := Error{
		Status:   http.StatusBadRequest,
		Type:     ErrorTypeBadPayload,
		Message:  ErrorMessageBadPayload,
		MoreInfo: "",
		Details:  details,
	}
	return &apiError
}
func ErrorResponseMissingFieldSource() (response *Error) {
	return createMissingFieldError(FieldSource)
}

func ErrorResponseMissingFieldSourceType() (response *Error) {
	return createMissingFieldError(FieldSourceType)
}

func ErrorResponseMissingFieldSourceNamespace() (response *Error) {
	return createMissingFieldError(FieldSourceNamespace)
}

func ErrorResponseMissingFieldSourceEnvironment() (response *Error) {
	return createMissingFieldError(FieldSourceEnvironment)
}

func ErrorResponseMissingFieldData() (response *Error) {
	return createMissingFieldError(FieldData)
}

func ErrorResponseMissingFieldEventType() (response *Error) {
	return createMissingFieldError(FieldEventType)
}

func ErrorResponseMissingFieldEventTypeVersion() (response *Error) {
	return createMissingFieldError(FieldEventTypeVersion)
}

func ErrorResponseMissingFieldEventTime() (response *Error) {
	return createMissingFieldError(FieldEventTime)
}

func ErrorResponseWrongSourceEnvironment() (response *Error) {
	return createInvalidFieldError(FieldSourceEnvironment)
}

func ErrorResponseWrongSourceNamespace() (response *Error) {
	return createInvalidFieldError(FieldSourceNamespace)
}

func ErrorResponseWrongSourceType() (response *Error) {
	return createInvalidFieldError(FieldSourceType)
}

func ErrorResponseWrongEventType() (response *Error) {
	return createInvalidFieldError(FieldEventType)
}

func ErrorResponseWrongEventTypeVersion() (response *Error) {
	return createInvalidFieldError(FieldEventTypeVersion)
}

func ErrorResponseWrongEventTime(err error) (response *Error) {
	return createInvalidFieldError(FieldEventTime)
}

func ErrorResponseWrongEventId() (response *Error) {
	return createInvalidFieldError(FieldEventId)
}

func createMissingFieldError(field interface{}) (response *Error) {
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

func createInvalidFieldError(field interface{}) (response *Error) {
	apiErrorDetail := ErrorDetail{
		Field:    field.(string),
		Type:     ErrorTypeInvalidField,
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
