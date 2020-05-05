package shared

import (
	"errors"
	"fmt"
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
	Verbose     bool
	Log         Logger
}

func LogReadiness(ready, verbose bool, name, namespace string, log Logger, resource interface{}) {
	typeName := fmt.Sprintf("%T", resource)

	if ready {
		log.Logf("%s %s in namespace %s is ready", typeName, name, namespace)

	} else {
		log.Logf("%s %s in namespace is not ready", typeName, name, namespace)
	}

	if verbose {
		log.Logf("%+v", resource)
	}
}
