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

func MakePluggableFunc(informer cache.SharedIndexInformer, pluggabilityEnabled bool) func(PluggableModule) {
	return func(module PluggableModule) {
		if !pluggabilityEnabled {
			err := module.Enable()
			printModuleErrorIfShould(err, module, "enabling")
			return
		}

		glog.Infof("Enabling module pluggability for %s", module.Name())
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
