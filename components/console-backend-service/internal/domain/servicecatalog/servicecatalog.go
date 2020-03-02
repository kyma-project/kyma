package servicecatalog

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	bindingApi "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	catalogInformers "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
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
	Find(namespace string, name string) (*bindingApi.ServiceBinding, error)
	ListForServiceInstance(namespace string, instanceName string) ([]*bindingApi.ServiceBinding, error)
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, rafterRetriever shared.RafterRetriever) (*PluggableContainer, error) {
	scCli, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing SC clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			scCli:                scCli,
			informerResyncPeriod: informerResyncPeriod,
			rafterRetriever:      rafterRetriever,
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
	scCli := r.cfg.scCli

	informerFactory := catalogInformers.NewSharedInformerFactory(scCli, informerResyncPeriod)

	r.informerFactory = informerFactory

	serviceInstanceService, err := newServiceInstanceService(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer(), scCli)
	if err != nil {
		return errors.Wrapf(err, "while creating service instance service")
	}
	servicePlanService, err := newServicePlanService(informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer(), r.cfg.rafterRetriever)
	if err != nil {
		return errors.Wrapf(err, "while creating service plan service")

	}
	serviceClassService, err := newServiceClassService(informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer())
	if err != nil {
		return errors.Wrapf(err, "while creating service class service")
	}
	serviceBrokerService := newServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer())
	serviceBindingService, err := newServiceBindingService(scCli.ServicecatalogV1beta1(), informerFactory.Servicecatalog().V1beta1().ServiceBindings().Informer(), name.Generate)
	if err != nil {
		return errors.Wrapf(err, "while creating service binding service")
	}

	clusterServiceClassService, err := newClusterServiceClassService(informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer())
	if err != nil {
		return errors.Wrapf(err, "while creating cluster service class service")
	}
	clusterServicePlanService, err := newClusterServicePlanService(informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer())
	if err != nil {
		return errors.Wrapf(err, "while creating cluster service plan service")
	}
	clusterServiceBrokerService := newClusterServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer())

	r.Pluggable.EnableAndSyncInformerFactory(r.informerFactory, func() {
		r.Resolver = &domainResolver{
			clusterServiceClassResolver:  newClusterServiceClassResolver(clusterServiceClassService, clusterServicePlanService, serviceInstanceService, r.cfg.rafterRetriever),
			serviceClassResolver:         newServiceClassResolver(serviceClassService, servicePlanService, serviceInstanceService, r.cfg.rafterRetriever),
			clusterServicePlanResolver:   newClusterServicePlanResolver(r.cfg.rafterRetriever),
			servicePlanResolver:          newServicePlanResolver(r.cfg.rafterRetriever),
			serviceInstanceResolver:      newServiceInstanceResolver(serviceInstanceService, clusterServicePlanService, clusterServiceClassService, servicePlanService, serviceClassService),
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
	scCli                clientset.Interface
	informerResyncPeriod time.Duration
	rafterRetriever      shared.RafterRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ClusterServiceClassQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error)
	ClusterServiceClassesQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error)
	ClusterServiceClassPlansField(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error)
	ClusterServiceClassInstancesField(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) ([]gqlschema.ServiceInstance, error)
	ClusterServiceClassActivatedField(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) (bool, error)
	ClusterServiceClassClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.ClusterAssetGroup, error)

	ServiceClassQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceClass, error)
	ServiceClassesQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceClass, error)
	ServiceClassPlansField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error)
	ServiceClassInstancesField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServiceInstance, error)
	ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error)
	ServiceClassClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.ClusterAssetGroup, error)
	ServiceClassAssetGroupField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.AssetGroup, error)

	ClusterServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServicePlan) (*gqlschema.ClusterAssetGroup, error)
	ClusterServicePlanAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.AssetGroup, error)

	CreateServiceInstanceMutation(ctx context.Context, namespace string, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error)
	DeleteServiceInstanceMutation(ctx context.Context, name, namespace string) (*gqlschema.ServiceInstance, error)
	ServiceInstanceQuery(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error)
	ServiceInstancesQuery(ctx context.Context, namespace string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error)
	ServiceInstanceEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceInstanceEvent, error)
	ServiceInstanceClusterServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error)
	ServiceInstanceClusterServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error)
	ServiceInstanceServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error)
	ServiceInstanceServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error)
	ServiceInstanceBindableField(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error)

	ClusterServiceBrokersQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceBroker, error)
	ClusterServiceBrokerQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error)
	ClusterServiceBrokerEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterServiceBrokerEvent, error)

	ServiceBrokersQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceBroker, error)
	ServiceBrokerQuery(ctx context.Context, name string, namespace string) (*gqlschema.ServiceBroker, error)
	ServiceBrokerEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBrokerEvent, error)

	CreateServiceBindingMutation(ctx context.Context, serviceBindingName *string, serviceInstanceName, env string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error)
	DeleteServiceBindingMutation(ctx context.Context, serviceBindingName, env string) (*gqlschema.DeleteServiceBindingOutput, error)
	ServiceBindingQuery(ctx context.Context, name, env string) (*gqlschema.ServiceBinding, error)
	ServiceBindingsToInstanceQuery(ctx context.Context, instanceName, namespace string) (*gqlschema.ServiceBindings, error)
	ServiceBindingEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBindingEvent, error)
}

type domainResolver struct {
	*clusterServiceClassResolver
	*serviceClassResolver
	*clusterServicePlanResolver
	*servicePlanResolver
	*serviceInstanceResolver
	*clusterServiceBrokerResolver
	*serviceBrokerResolver
	*serviceBindingResolver
}
