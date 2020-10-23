package shared

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

type Container struct {
	DynamicCli  dynamic.Interface
	Namespace   string
	WaitTimeout time.Duration
	Verbose     bool
	Log         *logrus.Entry
}

func (c Container) WithLogger(l *logrus.Entry) Container {
	c.Log = l
	return c
}

func LogReadiness(ready, verbose bool, name string, log *logrus.Entry, resource interface{}) {
	typeName := fmt.Sprintf("%T", resource)

	if ready {
		log.Infof("%s %s is ready", typeName, name)
	} else {
		log.Infof("%s %s is not ready", typeName, name)
	}

	if verbose {
		log.Infof("%+v", resource)
	}
}
