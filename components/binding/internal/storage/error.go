package storage

import (
	"fmt"

	"github.com/pkg/errors"
)

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
func (NotFound) NotFound() bool  { return true }

func IsNotFound(err error) bool {
	cause := errors.Cause(err)
	nfe, ok := cause.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}
