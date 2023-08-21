package utils

import (
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
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
