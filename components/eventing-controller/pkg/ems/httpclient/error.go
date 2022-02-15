package httpclient

import (
	"fmt"
	"strings"
)

type Error struct {
	StatusCode int
	Message    string
	Cause      error
}

type ErrorOpt func(*Error)

func NewError(err error, opts ...ErrorOpt) *Error {
	e := &Error{Cause: err}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Error provides a description of the error including all fields which are initialized to a non-default value
func (e Error) Error() string {
	messageParts := make([]string, 0)

	if e.Message != "" {
		messageDescription := fmt.Sprintf("message: %s", e.Message)
		messageParts = append(messageParts, messageDescription)
	}
	if e.StatusCode != 0 {
		statusCodeDescription := fmt.Sprintf("status code: %d", e.StatusCode)
		messageParts = append(messageParts, statusCodeDescription)
	}
	if e.Cause != nil {
		causeDescription := fmt.Sprintf("cause: %s", e.Cause.Error())
		messageParts = append(messageParts, causeDescription)
	}
	return strings.Join(messageParts, "; ")
}

func (e Error) Unwrap() error {
	return e.Cause
}

func WithStatusCode(statusCode int) ErrorOpt {
	return func(e *Error) {
		e.StatusCode = statusCode
	}
}

func WithMessage(message string) ErrorOpt {
	return func(e *Error) {
		e.Message = message
	}
}
