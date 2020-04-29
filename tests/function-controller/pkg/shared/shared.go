package shared

import (
	"errors"
)

type Logger interface {
	Logf(format string, args ...interface{})
}

var (
	ErrInvalidDataType = errors.New("invalid data type")
)
