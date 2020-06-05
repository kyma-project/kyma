package module

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=PluggableModule -output=automock -outpkg=automock -case=underscore
type PluggableModule interface {
	Enable() error
	Disable() error
	IsEnabled() bool
	Name() string
}

func MakePluggableFunc(informer cache.SharedIndexInformer) func(PluggableModule) {
	return func(module PluggableModule) {
		glog.Infof("Making the '%s' module pluggable...", module.Name())
		eventHandler := newEventHandler(module)
		informer.AddEventHandler(eventHandler)
	}
}

func printModuleErrorIfShould(err error, module PluggableModule, operationType string) {
	if err == nil {
		return
	}
	glog.Error(errors.Wrapf(err, "while %s module %s", operationType, module.Name()))
}
