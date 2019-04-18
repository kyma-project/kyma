package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	fakeSbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Addons Configurations

func NewAddonsConfigurationService(informer cache.SharedIndexInformer, cfgMapClient v1.ConfigMapInterface) *addonsConfigurationService {
	return newAddonsConfigurationService(informer, cfgMapClient)
}

func NewAddonsConfigurationConverter() *addonsConfigurationConverter {
	return &addonsConfigurationConverter{}
}

func NewAddonsConfigurationResolver(addonsCfgUpdater addonsCfgUpdater, addonsCfgMutations addonsCfgMutations, addonsCfgLister addonsCfgLister) *addonsConfigurationResolver {
	return &addonsConfigurationResolver{
		addonsCfgConverter: addonsConfigurationConverter{},
		addonsCfgUpdater:   addonsCfgUpdater,
		addonsCfgMutations: addonsCfgMutations,
		addonsCfgLister:    addonsCfgLister,
	}
}

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
	r.cfg.k8sClient = fake.NewSimpleClientset()
	r.cfg.dynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
}
