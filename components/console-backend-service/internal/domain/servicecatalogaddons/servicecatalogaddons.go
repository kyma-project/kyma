package servicecatalogaddons

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/experimental"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
	addonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned"
	addonsInformers "github.com/kyma-project/kyma/components/helm-broker/pkg/client/informers/externalversions"
	bindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bindingUsageClientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	bindingUsageInformers "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver                      Resolver
	ServiceCatalogAddonsRetriever *serviceCatalogAddonsRetriever
	sbuInformerFactory            bindingUsageInformers.SharedInformerFactory
	addonsInformerFactory         addonsInformers.SharedInformerFactory
	cmInformerFactory             v1.SharedInformerFactory
}

type serviceCatalogAddonsRetriever struct {
	ServiceBindingUsageLister ServiceBindingUsageLister
}

func (r *serviceCatalogAddonsRetriever) ServiceBindingUsage() shared.ServiceBindingUsageLister {
	return r.ServiceBindingUsageLister
}

//go:generate failery -name=ServiceBindingUsageLister -case=underscore -output disabled -outpkg disabled
type ServiceBindingUsageLister interface {
	ListForDeployment(namespace, kind, deploymentName string) ([]*bindingUsageApi.ServiceBindingUsage, error)
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, scRetriever shared.ServiceCatalogRetriever, featureToggles experimental.FeatureToggles) (*PluggableContainer, error) {
	k8sCli, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing k8s Clientset")
	}

	serviceBindingUsageClient, err := bindingUsageClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Binding Usage Clientset")
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	var addonsCfgCli addonsClientset.Interface
	if featureToggles.ClusterAddonsConfigurationCRDEnabled {
		addonsCfgCli, err = addonsClientset.NewForConfig(restConfig)
		if err != nil {
			return nil, errors.Wrap(err, "while initializing Addons Configuration Clientset")
		}
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			addonsCfgCli:                      addonsCfgCli,
			k8sClient:                         k8sCli,
			serviceBindingUsageClient:         serviceBindingUsageClient,
			dynamicClient:                     dynamicClient,
			informerResyncPeriod:              informerResyncPeriod,
			scRetriever:                       scRetriever,
			addonsConfigurationFeatureEnabled: featureToggles.ClusterAddonsConfigurationCRDEnabled,
		},
		ServiceCatalogAddonsRetriever: &serviceCatalogAddonsRetriever{},
		Pluggable:                     module.NewPluggable("servicecatalogaddons"),
	}
	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *PluggableContainer) Enable() error {
	informerResyncPeriod := r.cfg.informerResyncPeriod
	serviceBindingUsageClient := r.cfg.serviceBindingUsageClient
	k8sCli := r.cfg.k8sClient

	r.sbuInformerFactory = bindingUsageInformers.NewSharedInformerFactory(serviceBindingUsageClient, informerResyncPeriod)
	usageKindService := newUsageKindService(serviceBindingUsageClient.ServicecatalogV1alpha1(), r.cfg.dynamicClient, r.sbuInformerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer())
	serviceBindingUsageService, err := newServiceBindingUsageService(serviceBindingUsageClient.ServicecatalogV1alpha1(), r.sbuInformerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages().Informer(), r.cfg.scRetriever, name.Generate)
	if err != nil {
		return errors.Wrap(err, "while creating service binding usage service")
	}

	informersToSync := []module.SharedInformerFactory{r.sbuInformerFactory}

	var addonsConfigurationService *clusterAddonsConfigurationService
	if r.cfg.addonsConfigurationFeatureEnabled {
		r.addonsInformerFactory = addonsInformers.NewSharedInformerFactory(r.cfg.addonsCfgCli, informerResyncPeriod)
		informersToSync = append(informersToSync, r.addonsInformerFactory)
		clusterAddonsInformer := r.addonsInformerFactory.Addons().V1alpha1().ClusterAddonsConfigurations().Informer()
		addonsConfigurationService = newClusterAddonsConfigurationService(nil, clusterAddonsInformer, nil, r.cfg.addonsCfgCli.AddonsV1alpha1())
	} else {
		r.cmInformerFactory = v1.NewSharedInformerFactoryWithOptions(k8sCli, informerResyncPeriod, v1.WithNamespace(systemNs), v1.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = fmt.Sprintf("%s=%s", addonsCfgLabelKey, addonsCfgLabelValue)
		}))
		informersToSync = append(informersToSync, r.cmInformerFactory)
		cmInformer := r.cmInformerFactory.Core().V1().ConfigMaps().Informer()
		addonsConfigurationService = newClusterAddonsConfigurationService(cmInformer, nil, k8sCli.CoreV1().ConfigMaps(systemNs), nil)
	}

	onSyncHook := func() {
		r.Resolver = &domainResolver{
			serviceBindingUsageResolver: newServiceBindingUsageResolver(serviceBindingUsageService),
			usageKindResolver:           newUsageKindResolver(usageKindService),
			bindableResourcesResolver:   newBindableResourcesResolver(usageKindService),
			addonsConfigurationResolver: newAddonsRepoResolver(addonsConfigurationService, r.cfg.addonsConfigurationFeatureEnabled),
		}
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = serviceBindingUsageService

	}

	r.Pluggable.EnableAndSyncInformerFactories(onSyncHook, informersToSync...)

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = disabled.NewServiceBindingUsageLister(disabledErr)
		r.sbuInformerFactory = nil
		r.addonsInformerFactory = nil

	})

	return nil
}

type resolverConfig struct {
	k8sClient                         kubernetes.Interface
	serviceBindingUsageClient         bindingUsageClientset.Interface
	dynamicClient                     dynamic.Interface
	scRetriever                       shared.ServiceCatalogRetriever
	informerResyncPeriod              time.Duration
	addonsCfgCli                      addonsClientset.Interface
	addonsConfigurationFeatureEnabled bool
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	CreateServiceBindingUsageMutation(ctx context.Context, namespace string, input *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error)
	DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error)
	ServiceBindingUsageQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceBindingUsage, error)
	ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName, env string) ([]gqlschema.ServiceBindingUsage, error)
	ServiceBindingUsageEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBindingUsageEvent, error)

	ListUsageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error)
	ListBindableResources(ctx context.Context, namespace string) ([]gqlschema.BindableResourcesOutputItem, error)

	AddonsConfigurationsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error)
	CreateAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	UpdateAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	DeleteAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error)
	AddAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
	RemoveAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
	AddonsConfigurationEventSubscription(ctx context.Context) (<-chan gqlschema.AddonsConfigurationEvent, error)
}

type domainResolver struct {
	*serviceBindingUsageResolver
	*usageKindResolver
	*bindableResourcesResolver
	*addonsConfigurationResolver
}
