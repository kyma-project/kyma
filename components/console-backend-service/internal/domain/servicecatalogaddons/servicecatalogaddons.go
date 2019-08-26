package servicecatalogaddons

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
	addonsClientset "github.com/kyma-project/helm-broker/pkg/client/clientset/versioned"
	addonsInformers "github.com/kyma-project/helm-broker/pkg/client/informers/externalversions"
	bindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bindingUsageClientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	bindingUsageInformers "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver                      Resolver
	ServiceCatalogAddonsRetriever *serviceCatalogAddonsRetriever
	sbuInformerFactory            bindingUsageInformers.SharedInformerFactory
	addonsInformerFactory         addonsInformers.SharedInformerFactory
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

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, scRetriever shared.ServiceCatalogRetriever) (*PluggableContainer, error) {
	serviceBindingUsageClient, err := bindingUsageClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Binding Usage Clientset")
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	addonsCfgCli, err := addonsClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Addons Configuration Clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			addonsCfgCli:              addonsCfgCli,
			serviceBindingUsageClient: serviceBindingUsageClient,
			dynamicClient:             dynamicClient,
			informerResyncPeriod:      informerResyncPeriod,
			scRetriever:               scRetriever,
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

	r.sbuInformerFactory = bindingUsageInformers.NewSharedInformerFactory(serviceBindingUsageClient, informerResyncPeriod)
	usageKindService := newUsageKindService(serviceBindingUsageClient.ServicecatalogV1alpha1(), r.cfg.dynamicClient, r.sbuInformerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer())
	serviceBindingUsageService, err := newServiceBindingUsageService(serviceBindingUsageClient.ServicecatalogV1alpha1(), r.sbuInformerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages().Informer(), r.cfg.scRetriever, name.Generate)
	if err != nil {
		return errors.Wrap(err, "while creating service binding usage service")
	}

	r.addonsInformerFactory = addonsInformers.NewSharedInformerFactory(r.cfg.addonsCfgCli, informerResyncPeriod)
	clusterAddonsInformer := r.addonsInformerFactory.Addons().V1alpha1().ClusterAddonsConfigurations().Informer()
	clusterAddonsConfigurationService := newClusterAddonsConfigurationService(clusterAddonsInformer, r.cfg.addonsCfgCli.AddonsV1alpha1())
	addonsInformer := r.addonsInformerFactory.Addons().V1alpha1().AddonsConfigurations().Informer()
	addonsConfigurationService := newAddonsConfigurationService(addonsInformer, r.cfg.addonsCfgCli.AddonsV1alpha1())

	onSyncHook := func() {
		r.Resolver = &domainResolver{
			serviceBindingUsageResolver:        newServiceBindingUsageResolver(serviceBindingUsageService),
			usageKindResolver:                  newUsageKindResolver(usageKindService),
			bindableResourcesResolver:          newBindableResourcesResolver(usageKindService),
			addonsConfigurationResolver:        newAddonsConfigurationResolver(addonsConfigurationService),
			clusterAddonsConfigurationResolver: newClusterAddonsConfigurationResolver(clusterAddonsConfigurationService),
		}
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = serviceBindingUsageService

	}

	r.Pluggable.EnableAndSyncInformerFactories(onSyncHook, r.sbuInformerFactory, r.addonsInformerFactory)

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
	serviceBindingUsageClient bindingUsageClientset.Interface
	dynamicClient             dynamic.Interface
	scRetriever               shared.ServiceCatalogRetriever
	informerResyncPeriod      time.Duration
	addonsCfgCli              addonsClientset.Interface
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

	AddonsConfigurationsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error)
	CreateAddonsConfiguration(ctx context.Context, name, namespace string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	UpdateAddonsConfiguration(ctx context.Context, name, namespace string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	DeleteAddonsConfiguration(ctx context.Context, name, namespace string) (*gqlschema.AddonsConfiguration, error)
	AddAddonsConfigurationURLs(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error)
	RemoveAddonsConfigurationURLs(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error)
	ResyncAddonsConfiguration(ctx context.Context, name, namespace string) (*gqlschema.AddonsConfiguration, error)
	AddonsConfigurationEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.AddonsConfigurationEvent, error)

	ClusterAddonsConfigurationsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error)
	CreateClusterAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	UpdateClusterAddonsConfiguration(ctx context.Context, name string, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	DeleteClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error)
	AddClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
	RemoveClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
	ResyncClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error)
	ClusterAddonsConfigurationEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterAddonsConfigurationEvent, error)
}

type domainResolver struct {
	*serviceBindingUsageResolver
	*usageKindResolver
	*bindableResourcesResolver
	*clusterAddonsConfigurationResolver
	*addonsConfigurationResolver
}
