package servicecatalogaddons

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
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
	informerFactory               bindingUsageInformers.SharedInformerFactory
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

	container := &PluggableContainer{
		cfg: &resolverConfig{
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
	dynamicClient := r.cfg.dynamicClient

	r.informerFactory = bindingUsageInformers.NewSharedInformerFactory(serviceBindingUsageClient, informerResyncPeriod)
	usageKindService := newUsageKindService(serviceBindingUsageClient.ServicecatalogV1alpha1(), dynamicClient, r.informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer())
	serviceBindingUsageService, err := newServiceBindingUsageService(serviceBindingUsageClient.ServicecatalogV1alpha1(), r.informerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages().Informer(), r.cfg.scRetriever, name.Generate)
	if err != nil {
		return errors.Wrap(err, "while creating service binding usage service")
	}

	r.Pluggable.EnableAndSyncInformerFactory(r.informerFactory, func() {
		r.Resolver = &domainResolver{
			serviceBindingUsageResolver: newServiceBindingUsageResolver(serviceBindingUsageService),
			usageKindResolver:           newUsageKindResolver(usageKindService),
			bindableResourcesResolver:   newBindableResourcesResolver(usageKindService),
		}
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = serviceBindingUsageService
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.ServiceCatalogAddonsRetriever.ServiceBindingUsageLister = disabled.NewServiceBindingUsageLister(disabledErr)
		r.informerFactory = nil
	})

	return nil
}

type resolverConfig struct {
	serviceBindingUsageClient bindingUsageClientset.Interface
	dynamicClient             dynamic.Interface
	scRetriever               shared.ServiceCatalogRetriever
	informerResyncPeriod      time.Duration
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
}

type domainResolver struct {
	*serviceBindingUsageResolver
	*usageKindResolver
	*bindableResourcesResolver
}
