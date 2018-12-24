package module

import (
	"github.com/pkg/errors"
	"log"
	"time"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

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
			glog.Error(errors.Wrapf(err, "while enabling module %s", module.Name()))
			return
		}

		glog.Infof("Enable module pluggability: %s", module.Name())
		store := informer.GetStore()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				resource, ok := obj.(*v1alpha1.BackendModule)
				if !ok {
					glog.Error("Incorrect item type: %T, should be: *BackendModule", resource)
					return
				}

				moduleName := module.Name()
				if resource.Name != moduleName {
					return
				}

				isEnabled := module.IsEnabled()
				if isEnabled {
					return
				}

				if resource.Name == name && !isEnabled {
					err := module.Enable()

				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				glog.Infof("UpdateFunc \n %+v \n\n %+v", oldObj, newObj)
				//TODO: Is there a need to handle update event?
			},
			DeleteFunc: func(obj interface{}) {
				glog.Infof("DeleteFunc %+v", obj)
				//TODO: Is the deleted object is already deleted from store, or it needs to be filtered?
				req, err := isModuleRequired(store, module.Name())
				if err != nil {
					log.Fatal("Update Module Informer", err)
				}
				if req {
					module.Enable()
				} else {
					module.Disable()
				}
			},
		})
	}
}

func isModuleRequired(store cache.Store, moduleName string) (bool, error) {
	_, exists, err := store.GetByKey(moduleName)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func NewInformerFactory(restConfig *rest.Config, informerResyncPeriod time.Duration) (externalversions.SharedInformerFactory, error) {
	clientset, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	factory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	return factory, nil
}