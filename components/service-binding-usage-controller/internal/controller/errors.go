package controller

import (
	"fmt"

	"github.com/pkg/errors"
)

// NotFoundError indicates situation when given resource was not found
type NotFoundError struct {
	msg string
}

func (e NotFoundError) Error() string { return fmt.Sprintf("not found error: %s", e.msg) }

// NotFound returns information if such error should be treated as not found
func (NotFoundError) NotFound() bool { return true }

// NewNotFoundError returns a new not found error with given message
func NewNotFoundError(format string, args ...interface{}) NotFoundError {
	return NotFoundError{msg: fmt.Sprintf(format, args...)}
}

// IsNotFoundError checks if given error is NotFound error
func IsNotFoundError(err error) bool {
	err = errors.Cause(err)
	nfe, ok := err.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}

// ConflictError indicates situation when conflict occurs
type ConflictError struct {
	ConflictingResource string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("Conflict Error for [%s]", e.ConflictingResource)
}

type processBindingUsageError struct {
	Reason, Message string
}

func (s *processBindingUsageError) Error() string {
	return fmt.Sprintf("Reason: %s, details: %s", s.Reason, s.Message)
}

func newProcessServiceBindingError(reason string, err error) *processBindingUsageError {
	return &processBindingUsageError{
		Reason:  reason,
		Message: err.Error(),
	}
}
