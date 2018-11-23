package module

import (
	"log"
	"time"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/uiapi.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/generated/clientset/versioned"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/generated/informers/externalversions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type PluggableModule interface {
	Enable() error
	Disable() error
	IsEnabled() bool
	Name() string
}

func isModuleRequired(store cache.Store, moduleName string) (bool, error) {
	_, exists, err := store.GetByKey(moduleName)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func MakePluggableFunc(informer cache.SharedIndexInformer) func(PluggableModule) {
	return func(module PluggableModule) {
		glog.Infof("Enable module pluggability: %s", module.Name())
		store := informer.GetStore()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				glog.Infof("Add event %+v", obj)
				resource, ok := obj.(*v1alpha1.Module)
				if !ok {
					log.Printf("Module Conversion not ok %+v", obj)
				}

				name := module.Name()
				isEnabled := module.IsEnabled()

				if resource.Name == name && !isEnabled {
					module.Enable()
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

func NewInformerFactory(restConfig *rest.Config, informerResyncPeriod time.Duration) (externalversions.SharedInformerFactory, error) {
	clientset, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	factory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	return factory, nil
}
