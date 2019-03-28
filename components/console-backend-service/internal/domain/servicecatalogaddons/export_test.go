package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	fakeSbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
)

// Binding usage

func NewServiceBindingUsageService(buInterface v1alpha1.ServicecatalogV1alpha1Interface, informer cache.SharedIndexInformer, scRetriever shared.ServiceCatalogRetriever, sbuName string) (*serviceBindingUsageService, error) {
	return newServiceBindingUsageService(buInterface, informer, scRetriever, func() string {
		return sbuName
	})
}

func NewServiceBindingUsageResolver(op serviceBindingUsageOperations) *serviceBindingUsageResolver {
	return newServiceBindingUsageResolver(op)
}

// Service Catalog Module

func (r *PluggableContainer) SetFakeClient() {
	r.cfg.serviceBindingUsageClient = fakeSbu.NewSimpleClientset()
	r.cfg.dynamicClient = fakeDynamic.NewSimpleDynamicClient(nil)
}
