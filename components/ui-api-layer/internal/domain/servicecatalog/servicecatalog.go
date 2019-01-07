package servicecatalog

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"

	bindingApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	catalogInformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/name"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver                Resolver
	ServiceCatalogRetriever *serviceCatalogRetriever
	informerFactory         catalogInformers.SharedInformerFactory
}

type serviceCatalogRetriever struct {
	ServiceBindingFinderLister ServiceBindingFinderLister
}

func (r *serviceCatalogRetriever) ServiceBinding() shared.ServiceBindingFinderLister {
	return r.ServiceBindingFinderLister
}

//go:generate failery -name=ServiceBindingFinderLister -case=underscore -output disabled -outpkg disabled
type ServiceBindingFinderLister interface {
	Find(env string, name string) (*bindingApi.ServiceBinding, error)
	ListForServiceInstance(env string, instanceName string) ([]*bindingApi.ServiceBinding, error)
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, contentRetriever shared.ContentRetriever) (*PluggableContainer, error) {
	client, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			client:               client,
			informerResyncPeriod: informerResyncPeriod,
			contentRetriever:     contentRetriever,
		},
		Pluggable:               module.NewPluggable("servicecatalog"),
		ServiceCatalogRetriever: &serviceCatalogRetriever{},
	}
	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *PluggableContainer) Enable() error {
	informerResyncPeriod := r.cfg.informerResyncPeriod
	client := r.cfg.client

	contentRetriever := r.cfg.contentRetriever

	informerFactory := catalogInformers.NewSharedInformerFactory(client, informerResyncPeriod)
	r.informerFactory = informerFactory

	serviceInstanceService := newServiceInstanceService(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer(), client)
	servicePlanService := newServicePlanService(informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer())
	serviceClassService := newServiceClassService(informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer())
	serviceBrokerService := newServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer())
	serviceBindingService := newServiceBindingService(client.ServicecatalogV1beta1(), informerFactory.Servicecatalog().V1beta1().ServiceBindings().Informer(), name.Generate)

	clusterServiceClassService := newClusterServiceClassService(informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer())
	clusterServicePlanService := newClusterServicePlanService(informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer())
	clusterServiceBrokerService := newClusterServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer())

	r.Pluggable.EnableAndSyncInformerFactory(r.informerFactory, func() {
		r.Resolver = &domainResolver{
			serviceInstanceResolver:      newServiceInstanceResolver(serviceInstanceService, clusterServicePlanService, clusterServiceClassService, servicePlanService, serviceClassService),
			clusterServiceClassResolver:  newClusterServiceClassResolver(clusterServiceClassService, clusterServicePlanService, serviceInstanceService, contentRetriever),
			serviceClassResolver:         newServiceClassResolver(serviceClassService, servicePlanService, serviceInstanceService, contentRetriever),
			clusterServiceBrokerResolver: newClusterServiceBrokerResolver(clusterServiceBrokerService),
			serviceBrokerResolver:        newServiceBrokerResolver(serviceBrokerService),
			serviceBindingResolver:       newServiceBindingResolver(serviceBindingService),
		}
		r.ServiceCatalogRetriever.ServiceBindingFinderLister = serviceBindingService
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.ServiceCatalogRetriever.ServiceBindingFinderLister = disabled.NewServiceBindingFinderLister(disabledErr)
		r.informerFactory = nil
	})

	return nil
}

type resolverConfig struct {
	client               clientset.Interface
	informerResyncPeriod time.Duration
	contentRetriever     shared.ContentRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ClusterServiceClassQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error)
	ClusterServiceClassesQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error)
	ClusterServiceClassPlansField(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error)
	ClusterServiceClassActivatedField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (bool, error)
	ClusterServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error)
	ClusterServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error)
	ClusterServiceClassContentField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error)

	ServiceClassQuery(ctx context.Context, name, environment string) (*gqlschema.ServiceClass, error)
	ServiceClassesQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceClass, error)
	ServiceClassPlansField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error)
	ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error)
	ServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error)
	ServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error)
	ServiceClassContentField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error)

	CreateServiceInstanceMutation(ctx context.Context, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error)
	DeleteServiceInstanceMutation(ctx context.Context, name, environment string) (*gqlschema.ServiceInstance, error)
	ServiceInstanceQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error)
	ServiceInstancesQuery(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error)
	ServiceInstanceEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceInstanceEvent, error)
	ServiceInstanceClusterServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error)
	ServiceInstanceClusterServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error)
	ServiceInstanceServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error)
	ServiceInstanceServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error)
	ServiceInstanceBindableField(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error)

	ClusterServiceBrokersQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceBroker, error)
	ClusterServiceBrokerQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error)
	ClusterServiceBrokerEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterServiceBrokerEvent, error)

	ServiceBrokersQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceBroker, error)
	ServiceBrokerQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceBroker, error)
	ServiceBrokerEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBrokerEvent, error)

	CreateServiceBindingMutation(ctx context.Context, serviceBindingName *string, serviceInstanceName, env string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error)
	DeleteServiceBindingMutation(ctx context.Context, serviceBindingName, env string) (*gqlschema.DeleteServiceBindingOutput, error)
	ServiceBindingQuery(ctx context.Context, name, env string) (*gqlschema.ServiceBinding, error)
	ServiceBindingsToInstanceQuery(ctx context.Context, instanceName, environment string) (gqlschema.ServiceBindings, error)
	ServiceBindingEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingEvent, error)
}

type domainResolver struct {
	*clusterServiceClassResolver
	*serviceClassResolver
	*serviceInstanceResolver
	*clusterServiceBrokerResolver
	*serviceBrokerResolver
	*serviceBindingResolver
}
