package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
)

// Addons Configurations

func NewAddonsConfigurationService(addonsCfgInformer cache.SharedIndexInformer, addonsCfgClient dynamic.NamespaceableResourceInterface) *addonsConfigurationService {
	return newAddonsConfigurationService(addonsCfgInformer, addonsCfgClient)
}

func NewClusterAddonsConfigurationService(addonsCfgInformer cache.SharedIndexInformer, addonsCfgClient dynamic.ResourceInterface) *clusterAddonsConfigurationService {
	return newClusterAddonsConfigurationService(addonsCfgInformer, addonsCfgClient)
}

func NewBindableResourcesResolver(lister bindableResourceLister) *bindableResourcesResolver {
	return newBindableResourcesResolver(lister)
}

func NewServiceBindingUsageConverter() *serviceBindingUsageConverter {
	return &serviceBindingUsageConverter{
		extractor: &status.BindingUsageExtractor{},
	}
}

func NewAddonsConfigurationConverter() *addonsConfigurationConverter {
	return &addonsConfigurationConverter{}
}

func NewClusterAddonsConfigurationConverter() *clusterAddonsConfigurationConverter {
	return &clusterAddonsConfigurationConverter{}
}

func NewClusterAddonsConfigurationResolver(addonsCfgUpdater clusterAddonsCfgUpdater, addonsCfgMutations clusterAddonsCfgMutations, addonsCfgLister clusterAddonsCfgLister) *clusterAddonsConfigurationResolver {
	return &clusterAddonsConfigurationResolver{
		addonsCfgConverter: clusterAddonsConfigurationConverter{},
		addonsCfgUpdater:   addonsCfgUpdater,
		addonsCfgMutations: addonsCfgMutations,
		addonsCfgLister:    addonsCfgLister,
	}
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

func NewServiceBindingUsageService(sbuClient dynamic.NamespaceableResourceInterface, informer cache.SharedIndexInformer, scRetriever shared.ServiceCatalogRetriever, sbuName string) (*serviceBindingUsageService, error) {
	return newServiceBindingUsageService(sbuClient, informer, scRetriever, func() string {
		return sbuName
	})
}

func NewServiceBindingUsageResolver(op serviceBindingUsageOperations, converter gqlServiceBindingUsageConverter) *serviceBindingUsageResolver {
	return newServiceBindingUsageResolver(op, converter)
}

func NewUsageKindResolver(svc usageKindServices) *usageKindResolver {
	return newUsageKindResolver(svc)
}

func NewUsageKindService(res dynamic.Interface, informer cache.SharedIndexInformer) *usageKindService {
	return newUsageKindService(res, informer)
}

func NewUsageKindConverter() *usageKindConverter {
	return &usageKindConverter{}
}

// Service Catalog Module
func (r *PluggableContainer) SetFakeClient() {
	r.cfg.dynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
}
