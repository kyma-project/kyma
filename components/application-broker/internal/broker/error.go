package broker

import "github.com/pkg/errors"

type notFoundError struct{}

func (notFoundError) Error() string  { return "not found" }
func (notFoundError) NotFound() bool { return true }

// IsNotFoundError check if error is NotFound one.
func IsNotFoundError(err error) bool {
	cause := errors.Cause(err)

	nfe, ok := cause.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}

// IsForbiddenError checks if error represent Forbidden one.
func IsForbiddenError(err error) bool {
	type forbidden interface {
		Forbidden() bool
	}

	if t, ok := err.(forbidden); ok {
		return t.Forbidden()
	}
	return false
}

// ForbiddenError represents situation when operation is forbidden
type ForbiddenError struct {
}

func (fe *ForbiddenError) Error() string {
	return "Forbidden Error"
}

// Forbidden is a marker method, used in IsForbiddenError method
func (fe *ForbiddenError) Forbidden() bool {
	return true
}
