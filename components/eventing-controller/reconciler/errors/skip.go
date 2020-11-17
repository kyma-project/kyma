package errors

import "golang.org/x/xerrors"

// skippable is an interface of error types that can be skipped.
type skippable interface {
	skip()
}

// IsSkippable returns true if the error is skippable, otherwise returns false.
func IsSkippable(err error) bool {
	if err == nil {
		return true
	}
	_, ok := err.(skippable)
	return ok
}

// skippableError is an implementation of a skippable reconcile error.
type skippableError struct {
	err error
}

var (
	// compile-time checks
	_ error           = &skippableError{}
	_ skippable       = &skippableError{}
	_ xerrors.Wrapper = &skippableError{}
)

// NewSkippable returns a new skippable error.
func NewSkippable(err error) error {
	return &skippableError{err: err}
}

// Error implements the error interface.
func (e *skippableError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

// skip implements the skippable interface.
func (*skippableError) skip() {}

// skip implements the Wrapper interface.
func (e *skippableError) Unwrap() error {
	return e.err
}
