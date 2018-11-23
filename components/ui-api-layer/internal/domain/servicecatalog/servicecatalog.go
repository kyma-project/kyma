package servicecatalog

import (
	"time"

	bindingApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	catalogInformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	bindingUsageApi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bindingUsageClientset "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	bindingUsageInformers "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/name"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Container struct {
	Resolver                  *Resolver
	ServiceBindingUsageLister ServiceBindingUsageLister
	ServiceBindingGetter      ServiceBindingGetter
}

type ServiceBindingUsageLister interface {
	ListForDeployment(environment, kind, deploymentName string) ([]*bindingUsageApi.ServiceBindingUsage, error)
}

type ServiceBindingGetter interface {
	Find(env string, name string) (*bindingApi.ServiceBinding, error)
}

type Resolver struct {
	*clusterServiceClassResolver
	*serviceClassResolver
	*serviceInstanceResolver
	*clusterServiceBrokerResolver
	*serviceBrokerResolver
	*serviceBindingResolver
	*serviceBindingUsageResolver
	*usageKindResolver
	*bindableResourcesResolver

	informerFactory             catalogInformers.SharedInformerFactory
	bindingUsageInformerFactory bindingUsageInformers.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, asyncApiSpecGetter AsyncApiSpecGetter, apiSpecGetter ApiSpecGetter, contentGetter ContentGetter) (*Container, error) {
	client, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Clientset")
	}

	informerFactory := catalogInformers.NewSharedInformerFactory(client, informerResyncPeriod)

	serviceInstanceService := newServiceInstanceService(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer(), client)
	servicePlanService := newServicePlanService(informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer())
	serviceClassService := newServiceClassService(informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer())
	serviceBrokerService := newServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer())
	serviceBindingService := newServiceBindingService(client.ServicecatalogV1beta1(), informerFactory.Servicecatalog().V1beta1().ServiceBindings().Informer(), name.Generate)

	clusterServiceClassService := newClusterServiceClassService(informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer())
	clusterServicePlanService := newClusterServicePlanService(informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer())
	clusterServiceBrokerService := newClusterServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer())

	serviceBindingUsageClient, err := bindingUsageClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Binding Usage Clientset")
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	serviceBindingUsageInformerFactory := bindingUsageInformers.NewSharedInformerFactory(serviceBindingUsageClient, informerResyncPeriod)
	usageKindService := newUsageKindService(serviceBindingUsageClient.ServicecatalogV1alpha1(), dynamicClient, serviceBindingUsageInformerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer())
	serviceBindingUsageService := newServiceBindingUsageService(serviceBindingUsageClient.ServicecatalogV1alpha1(), serviceBindingUsageInformerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages().Informer(), serviceBindingService, name.Generate)

	return &Container{
		Resolver: &Resolver{
			informerFactory:              informerFactory,
			bindingUsageInformerFactory:  serviceBindingUsageInformerFactory,
			serviceInstanceResolver:      newServiceInstanceResolver(serviceInstanceService, clusterServicePlanService, clusterServiceClassService, servicePlanService, serviceClassService),
			clusterServiceClassResolver:  newClusterServiceClassResolver(clusterServiceClassService, clusterServicePlanService, serviceInstanceService, asyncApiSpecGetter, apiSpecGetter, contentGetter),
			serviceClassResolver:         newServiceClassResolver(serviceClassService, servicePlanService, serviceInstanceService, asyncApiSpecGetter, apiSpecGetter, contentGetter),
			clusterServiceBrokerResolver: newClusterServiceBrokerResolver(clusterServiceBrokerService),
			serviceBrokerResolver:        newServiceBrokerResolver(serviceBrokerService),
			serviceBindingResolver:       newServiceBindingResolver(serviceBindingService),
			serviceBindingUsageResolver:  newServiceBindingUsageResolver(serviceBindingUsageService),
			usageKindResolver:            newUsageKindResolver(usageKindService),
			bindableResourcesResolver:    newBindableResourcesResolver(usageKindService),
		},
		ServiceBindingUsageLister: serviceBindingUsageService,
		ServiceBindingGetter:      serviceBindingService,
	}, nil
}

//func (r *Resolver) Enable() {
//
//}
//func (r *Resolver) Disable() {
//
//}
//func (r *Resolver) IsEnabled() bool {
//
//}
//func (r *Resolver) Name() string {
//
//}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
	r.bindingUsageInformerFactory.Start(stopCh)
	r.bindingUsageInformerFactory.WaitForCacheSync(stopCh)
}
