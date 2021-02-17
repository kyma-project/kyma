package error

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
