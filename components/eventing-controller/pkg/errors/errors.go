package errors

import (
	"errors"
	"fmt"
)

// MakeError creates a new error and includes the underlyingError in the message.
// However, it does not expose/wrap the underlyingError.
//
// Following the recommendation of https://go.dev/blog/go1.13-errors,
// the actualError is encapsulated into a new error and not returned directly.
// This forces callers to use
// errors.Is(err, pkg.ErrPermission) instead of
// err == pkg.ErrPermission { … }.
func MakeError(actualError, underlyingError error) error {
	return fmt.Errorf("%w: %v", actualError, underlyingError)
}

// ArgumentError is a generic error which can be used to create errors that shall include any kind of argument
// in their Error() message.
// By using the ArgumentError you don't need to create a custom error for each of these errors that have this
// custom argument.
// Instead, a custom error can be defined as a variable once.
// The equality of two ArgumentError is defined through the equality of errorFormatType.
// errorFormatType is supposed to be a format string so that the argument
// provided with WithArg can be printed in Error().
//
// See ExampleArgumentError for an example.
type ArgumentError struct {
	errorFormatType string
	argument        string
}

// NewArgumentError creates a new ArgumentError.
func NewArgumentError(errorFormatType string) *ArgumentError {
	return &ArgumentError{errorFormatType: errorFormatType}
}

// Error returns a human-readable string representation of the error.
func (e *ArgumentError) Error() string {
	return fmt.Sprintf(e.errorFormatType, e.argument)
}

// WithArg shall be used to provide the argument which will be printed according to the format string
// provided in errorFormatType via NewArgumentError.
func (e *ArgumentError) WithArg(argument string) *ArgumentError {
	return &ArgumentError{
		errorFormatType: e.errorFormatType,
		argument:        argument,
	}
}

// Is defines the equality of an ArgumentError. Two ArgumentError are equal if their errorFormatType is equal.
// Use as following:
//
// var errInvalidStorageType = ArgumentError{errorFormatType: "invalid stream storage type: %q"}
// ok := errors.Is(err, &errInvalidStorageType).
func (e *ArgumentError) Is(target error) bool {
	var genericError *ArgumentError
	ok := errors.As(target, &genericError)
	if !ok {
		return false
	}
	// errors are considered equal if their errorFormatType is equal!
	return e.errorFormatType == genericError.errorFormatType
}
