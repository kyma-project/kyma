package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"k8s.io/client-go/tools/cache"
)

const (
	namespaceExternalIDIndexName = "namespaceExtID"
)

// Facade expose operations on ServiceInstance objects
type Facade struct {
	siInformer cache.SharedIndexInformer
	scInformer cache.SharedIndexInformer
}

// NewFacade creates Service Catalog Facade
func NewFacade(informer cache.SharedIndexInformer, classInformer cache.SharedIndexInformer) *Facade {
	informer.AddIndexers(cache.Indexers{
		namespaceExternalIDIndexName: func(obj interface{}) ([]string, error) {
			si, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceInstance", obj, obj)
			}

			return []string{namespaceExtIDKey(si.Namespace, si.Spec.ExternalID)}, nil
		},
	})

	return &Facade{
		siInformer: informer,
		scInformer: classInformer,
	}
}

// AnyServiceInstanceExists checks whether there is at least one service instance created application service class
func (f *Facade) AnyServiceInstanceExists(namespace string) (bool, error) {
	instances, err := f.siInformer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)
	if err != nil {
		return false, err
	}

	for _, item := range instances {
		instance, ok := item.(*v1beta1.ServiceInstance)
		if !ok {
			return false, fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceInstance", item, item)
		}
		if instance.Spec.ServiceClassRef == nil {
			continue
		}
		clKey := fmt.Sprintf("%s/%s", namespace, instance.Spec.ServiceClassRef.Name)
		clObj, exists, err := f.scInformer.GetStore().GetByKey(clKey)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}

		cl, ok := clObj.(*v1beta1.ServiceClass)
		if !ok {
			return false, fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceClass", clObj, clObj)
		}
		if cl.Spec.ServiceBrokerName == nsbroker.NamespacedBrokerName {
			return true, nil
		}
	}

	return false, nil
}

func namespaceExtIDKey(namespace string, extID string) string {
	return fmt.Sprintf("%s/%s", namespace, extID)
}
