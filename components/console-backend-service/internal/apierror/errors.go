package apierror

import (
	"fmt"
)

type APIError struct {
	status  Status
	message string
}

var _ error = &APIError{}

func (e *APIError) Error() string {
	return e.message
}

func (e *APIError) Status() Status {
	return e.status
}

type ErrorField string

type ErrorFieldAggregate []ErrorField

func (agg ErrorFieldAggregate) String() string {
	if len(agg) == 0 {
		return ""
	}
	if len(agg) == 1 {
		return fmt.Sprintf("%s", agg[0])
	}
	result := fmt.Sprintf("[%s", agg[0])
	for i := 1; i < len(agg); i++ {
		result += fmt.Sprintf(", %s", agg[i])
	}
	result += "]"
	return result
}

type Status int

const (
	Unknown Status = iota
	Invalid
)

func NewInvalid(kind fmt.Stringer, errs ErrorFieldAggregate) *APIError {
	message := ""
	if kind != nil {
		message += fmt.Sprintf("%s ", kind)
	} else {
		message += "Resource "
	}
	message += "is invalid"
	if len(errs) > 0 {
		message += fmt.Sprintf(": %v", errs)
	}

	return &APIError{
		status:  Invalid,
		message: message,
	}
}

func NewInvalidField(path string, value string, detail string) ErrorField {
	message := ""
	if len(path) > 0 {
		message += fmt.Sprintf("`%s` ", path)
	}
	message += "field "
	if len(value) > 0 {
		message += fmt.Sprintf("(%s) ", value)
	}
	message += "is invalid"
	if len(detail) != 0 {
		message += fmt.Sprintf(": %s", detail)
	}
	return ErrorField(message)
}

func NewMissingField(path string) ErrorField {
	message := ""
	if len(path) > 0 {
		message += fmt.Sprintf("`%s` ", path)
	}
	message += "field is missing"
	return ErrorField(message)
}

func IsInvalid(err error) bool {
	return statusForError(err) == Invalid
}

func statusForError(err error) Status {
	type errorWithStatus interface {
		Status() Status
	}

	switch t := err.(type) {
	case errorWithStatus:
		return t.Status()
	}

	return Unknown
}
