package broker

import (
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

const namespaceExternalIDIndexName = "namespaceExtID"

// ServiceInstanceFacade expose operations on ServiceInstance objects
type ServiceInstanceFacade struct {
	informer cache.SharedIndexInformer
}

// NewServiceInstanceFacade creates ServiceInstanceFacade
func NewServiceInstanceFacade(informer cache.SharedIndexInformer) *ServiceInstanceFacade {
	informer.AddIndexers(cache.Indexers{
		namespaceExternalIDIndexName: func(obj interface{}) ([]string, error) {
			si, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceInstance", obj, obj)
			}

			return []string{namespaceExtIDKey(si.Namespace, si.Spec.ExternalID)}, nil
		},
	})
	return &ServiceInstanceFacade{
		informer: informer,
	}
}

// GetByNamespaceAndExternalID returns service instance
func (f *ServiceInstanceFacade) GetByNamespaceAndExternalID(namespace string, extID string) (*v1beta1.ServiceInstance, error) {
	values, err := f.informer.GetIndexer().ByIndex(namespaceExternalIDIndexName, namespaceExtIDKey(namespace, extID))
	if err != nil {
		return nil, errors.Wrapf(err, "while getting service instance [namespace: %q ExtID: %q]", namespace, extID)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("service instance not found [namespace: %q ExtID: %q]", namespace, extID)
	}
	if len(values) > 1 {
		return nil, fmt.Errorf("more than one service instance found in namespace: %q with ExtID: %q", namespace, extID)
	}

	si, ok := values[0].(*v1beta1.ServiceInstance)
	if !ok {
		return nil, fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceInstance", values[0], values[0])
	}
	return si, nil
}

func namespaceExtIDKey(namespace string, extID string) string {
	return fmt.Sprintf("%s/%s", namespace, extID)
}
