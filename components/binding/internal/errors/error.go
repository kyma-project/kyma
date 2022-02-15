package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type TemporaryError struct {
	message string
}

func NewTemporaryError(msg string, args ...interface{}) *TemporaryError {
	return &TemporaryError{
		message: fmt.Sprintf(msg, args...),
	}
}

func AsTemporaryError(err error, context string, args ...interface{}) *TemporaryError {
	errCtx := fmt.Sprintf(context, args...)
	msg := fmt.Sprintf("%s: %s", errCtx, err.Error())

	return NewTemporaryError(msg)
}

func (te TemporaryError) Error() string { return te.message }
func (TemporaryError) Temporary() bool  { return true }

func IsTemporaryError(err error) bool {
	cause := errors.Cause(err)
	nfe, ok := cause.(interface {
		Temporary() bool
	})
	return ok && nfe.Temporary()
}

type NotFound struct {
	message string
}

func NewNotFound(msg string, args ...interface{}) *NotFound {
	return &NotFound{
		message: fmt.Sprintf(msg, args...),
	}
}

func AsNotFound(err error, context string, args ...interface{}) *NotFound {
	errCtx := fmt.Sprintf(context, args...)
	msg := fmt.Sprintf("%s: %s", errCtx, err.Error())

	return NewNotFound(msg)
}

func (te NotFound) Error() string { return te.message }
func (NotFound) NotFound() bool   { return true }

func IsNotFound(err error) bool {
	cause := errors.Cause(err)
	nfe, ok := cause.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}
