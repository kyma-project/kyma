package servicecatalogaddons

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	fakeaddonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/fake"
	addonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/typed/addons/v1alpha1"
	fakeSbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Addons Configurations

func NewAddonsConfigurationService(cmInformer cache.SharedIndexInformer, addonsCfgInformer cache.SharedIndexInformer, cfgMapClient v1.ConfigMapInterface, addonsCfgClient addonsClientset.AddonsV1alpha1Interface) *clusterAddonsConfigurationService {
	return newClusterAddonsConfigurationService(cmInformer, addonsCfgInformer, cfgMapClient, addonsCfgClient)
}

func NewAddonsConfigurationConverter() *clusterAddonsConfigurationConverter {
	return &clusterAddonsConfigurationConverter{}
}

func NewAddonsConfigurationResolver(addonsCfgUpdater addonsCfgUpdater, addonsCfgMutations addonsCfgMutations, addonsCfgLister addonsCfgLister, addonsConfigurationFeatureEnabled bool) *addonsConfigurationResolver {
	return &addonsConfigurationResolver{
		addonsCfgConverter:                clusterAddonsConfigurationConverter{},
		addonsCfgUpdater:                  addonsCfgUpdater,
		addonsCfgMutations:                addonsCfgMutations,
		addonsCfgLister:                   addonsCfgLister,
		addonsConfigurationFeatureEnabled: addonsConfigurationFeatureEnabled,
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
	r.cfg.addonsCfgCli = fakeaddonsClientset.NewSimpleClientset()
	r.cfg.serviceBindingUsageClient = fakeSbu.NewSimpleClientset()
	r.cfg.k8sClient = fake.NewSimpleClientset()
	r.cfg.dynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
}
