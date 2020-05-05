package shared

import (
	"errors"
	"time"

	"k8s.io/client-go/dynamic"
)

type Logger interface {
	Logf(format string, args ...interface{})
}

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

type Container struct {
	DynamicCli  dynamic.Interface
	Namespace   string
	WaitTimeout time.Duration
	Kind        string
	Verbose     bool
	Log         Logger
}
