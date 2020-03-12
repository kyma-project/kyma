package servicecatalog

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/client-go/tools/cache"
)

const (
	serviceBindingExternalIDIndexName = "namespaceExtID"
)

// ServiceBindingFetcher provides functionality to fetch ServiceBinding based on its external id
type ServiceBindingFetcher struct {
	sbInformer cache.SharedIndexInformer
}

// NewServiceBindingFetcher returns new instance of ServiceBindingFetcher
func NewServiceBindingFetcher(sbInformer cache.SharedIndexInformer) *ServiceBindingFetcher {
	sbInformer.AddIndexers(cache.Indexers{
		serviceBindingExternalIDIndexName: func(obj interface{}) ([]string, error) {
			sb, ok := obj.(*v1beta1.ServiceBinding)
			if !ok {
				return nil, fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceBinding", obj, obj)
			}

			return []string{namespaceExtIDKey(sb.Namespace, sb.Spec.ExternalID)}, nil
		},
	})

	return &ServiceBindingFetcher{
		sbInformer: sbInformer,
	}
}

// AnyServiceInstanceExists checks whether there is at least one service instance created application service class
func (f *ServiceBindingFetcher) GetServiceBindingSecretName(ns, externalID string) (string, error) {
	bindings, err := f.sbInformer.GetIndexer().ByIndex(serviceBindingExternalIDIndexName, namespaceExtIDKey(ns, externalID))
	if err != nil {
		return "", errors.Wrap(err, "while getting ServiceBinding from cache")
	}

	if len(bindings) != 1 {
		return "", errors.Errorf("expected to found one Service Binding but got %d", len(bindings))
	}
	item := bindings[0]
	sb, ok := item.(*v1beta1.ServiceBinding)
	if !ok {
		return "", fmt.Errorf("cannot covert object [%+v] of type %T to *v1beta1.ServiceBinding", item, item)
	}
	return sb.Spec.SecretName, nil
}
