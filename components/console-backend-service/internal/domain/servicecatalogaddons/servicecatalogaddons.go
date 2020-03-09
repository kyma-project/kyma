package servicecatalogaddons

import (
	"context"
	"time"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	bindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver                      Resolver
	ServiceCatalogAddonsRetriever *serviceCatalogAddonsRetriever
	informerFactory               dynamicinformer.DynamicSharedInformerFactory
}

type serviceCatalogAddonsRetriever struct {
	ServiceBindingUsageLister       ServiceBindingUsageLister
	GqlServiceBindingUsageConverter GqlServiceBindingUsageConverter
}

func (r *serviceCatalogAddonsRetriever) ServiceBindingUsage() shared.ServiceBindingUsageLister {
	return r.ServiceBindingUsageLister
}

func (r *serviceCatalogAddonsRetriever) ServiceBindingUsageConverter() shared.GqlServiceBindingUsageConverter {
	return r.GqlServiceBindingUsageConverter
}

//go:generate failery -name=ServiceBindingUsageLister -case=underscore -output disabled -outpkg disabled
type ServiceBindingUsageLister interface {
	ListByUsageKind(namespace, kind, resourceName string) ([]*bindingUsageApi.ServiceBindingUsage, error)
}

//go:generate failery -name=GqlServiceBindingUsageConverter -case=underscore -output disabled -outpkg disabled
type GqlServiceBindingUsageConverter interface {
	ToGQL(item *bindingUsageApi.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error)
	ToGQLs(in []*bindingUsageApi.ServiceBindingUsage) ([]gqlschema.ServiceBindingUsage, error)
}

var (
	usageKindsGVR = schema.GroupVersionResource{
		Version:  bindingUsageApi.SchemeGroupVersion.Version,
		Group:    bindingUsageApi.SchemeGroupVersion.Group,
		Resource: "usagekinds",
	}
	bindingUsageGVR = schema.GroupVersionResource{
		Version:  bindingUsageApi.SchemeGroupVersion.Version,
		Group:    bindingUsageApi.SchemeGroupVersion.Group,
		Resource: "servicebindingusages",
	}
	addonsConfigGVR = schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "addonsconfigurations",
	}
	clusterAddonsConfigGVR = schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clusteraddonsconfigurations",
	}
)

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, scRetriever shared.ServiceCatalogRetriever) (*PluggableContainer, error) {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			dynamicClient:        dynamicClient,
			informerResyncPeriod: informerResyncPeriod,
			scRetriever:          scRetriever,
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
	r.informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(r.cfg.dynamicClient, r.cfg.informerResyncPeriod)

	usageKindService := newUsageKindService(r.cfg.dynamicClient, r.informerFactory.ForResource(usageKindsGVR).Informer())

	serviceBindingUsageService, err := newServiceBindingUsageService(r.cfg.dynamicClient.Resource(bindingUsageGVR), r.informerFactory.ForResource(bindingUsageGVR).Informer(), r.cfg.scRetriever, name.Generate)
	if err != nil {
		return errors.Wrap(err, "while creating service binding usage service")
	}
	serviceBindingUsageConverter := newBindingUsageConverter()

	clusterAddonsConfigurationService := newClusterAddonsConfigurationService(r.informerFactory.ForResource(clusterAddonsConfigGVR).Informer(), r.cfg.dynamicClient.Resource(clusterAddonsConfigGVR))
	addonsConfigurationService := newAddonsConfigurationService(r.informerFactory.ForResource(addonsConfigGVR).Informer(), r.cfg.dynamicClient.Resource(addonsConfigGVR))

	onSyncHook := func() {
		r.Resolver = &domainResolver{
			serviceBindingUsageResolver:        newServiceBindingUsageResolver(serviceBindingUsageService, &serviceBindingUsageConverter),
			usageKindResolver:                  newUsageKindResolver(usageKindService),
			bindableResourcesResolver:          newBindableResourcesResolver(usageKindService),
			addonsConfigurationResolver:        newAddonsConfigurationResolver(addonsConfigurationService),
			clusterAddonsConfigurationResolver: newClusterAddonsConfigurationResolver(clusterAddonsConfigurationService),
		}
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = serviceBindingUsageService
		r.ServiceCatalogAddonsRetriever.GqlServiceBindingUsageConverter = &serviceBindingUsageConverter
	}

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.informerFactory, onSyncHook)
	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = disabled.NewServiceBindingUsageLister(disabledErr)
		r.ServiceCatalogAddonsRetriever.GqlServiceBindingUsageConverter = disabled.NewGqlServiceBindingUsageConverter(disabledErr)
	})

	return nil
}

type resolverConfig struct {
	dynamicClient        dynamic.Interface
	scRetriever          shared.ServiceCatalogRetriever
	informerResyncPeriod time.Duration
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	CreateServiceBindingUsageMutation(ctx context.Context, namespace string, input *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error)
	DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error)
	DeleteServiceBindingUsagesMutation(ctx context.Context, serviceBindingUsageNames []string, namespace string) ([]*gqlschema.DeleteServiceBindingUsageOutput, error)
	ServiceBindingUsageQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceBindingUsage, error)
	ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName, env string) ([]gqlschema.ServiceBindingUsage, error)
	ServiceBindingUsageEventSubscription(ctx context.Context, namespace string, resourceKind, resourceName *string) (<-chan gqlschema.ServiceBindingUsageEvent, error)

	ListUsageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error)
	ListBindableResources(ctx context.Context, namespace string) ([]gqlschema.BindableResourcesOutputItem, error)

	AddonsConfigurationsQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error)
	CreateAddonsConfiguration(ctx context.Context, name, namespace string, repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	UpdateAddonsConfiguration(ctx context.Context, name, namespace string, repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	DeleteAddonsConfiguration(ctx context.Context, name, namespace string) (*gqlschema.AddonsConfiguration, error)
	AddAddonsConfigurationURLs(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error)
	RemoveAddonsConfigurationURLs(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error)
	AddAddonsConfigurationRepositories(ctx context.Context, name, namespace string, repositories []gqlschema.AddonsConfigurationRepositoryInput) (*gqlschema.AddonsConfiguration, error)
	RemoveAddonsConfigurationRepositories(ctx context.Context, name, namespace string, urls []string) (*gqlschema.AddonsConfiguration, error)
	ResyncAddonsConfiguration(ctx context.Context, name, namespace string) (*gqlschema.AddonsConfiguration, error)
	AddonsConfigurationEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.AddonsConfigurationEvent, error)

	ClusterAddonsConfigurationsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.AddonsConfiguration, error)
	CreateClusterAddonsConfiguration(ctx context.Context, name string, repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	UpdateClusterAddonsConfiguration(ctx context.Context, name string, repositories []gqlschema.AddonsConfigurationRepositoryInput, urls []string, labels *gqlschema.Labels) (*gqlschema.AddonsConfiguration, error)
	DeleteClusterAddonsConfiguration(ctx context.Context, name string) (*gqlschema.AddonsConfiguration, error)
	AddClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
	RemoveClusterAddonsConfigurationURLs(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
	AddClusterAddonsConfigurationRepositories(ctx context.Context, name string, repositories []gqlschema.AddonsConfigurationRepositoryInput) (*gqlschema.AddonsConfiguration, error)
	RemoveClusterAddonsConfigurationRepositories(ctx context.Context, name string, urls []string) (*gqlschema.AddonsConfiguration, error)
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
