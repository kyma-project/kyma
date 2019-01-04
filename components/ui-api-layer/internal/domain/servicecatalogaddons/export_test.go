package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

// Binding usage

func NewServiceBindingUsageService(buInterface v1alpha1.ServicecatalogV1alpha1Interface, informer cache.SharedIndexInformer, bindingOp serviceBindingOperations, sbuName string) *serviceBindingUsageService {
	return newServiceBindingUsageService(buInterface, informer, bindingOp, func() string {
		return sbuName
	})
}

func NewServiceBindingUsageResolver(op serviceBindingUsageOperations) *serviceBindingUsageResolver {
	return newServiceBindingUsageResolver(op)
}
