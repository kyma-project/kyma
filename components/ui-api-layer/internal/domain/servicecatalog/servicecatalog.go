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

	instanceService := newServiceInstanceService(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer(), client)

	clusterServicePlanService := newClusterServicePlanService(informerFactory.Servicecatalog().V1beta1().ClusterServicePlans().Informer())
	servicePlanService := newServicePlanService(informerFactory.Servicecatalog().V1beta1().ServicePlans().Informer())

	clusterServiceClassService := newClusterServiceClassService(informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer())
	serviceClassService := newServiceClassService(informerFactory.Servicecatalog().V1beta1().ServiceClasses().Informer())

	clusterServiceBrokerService := newClusterServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ClusterServiceBrokers().Informer())
	serviceBrokerService := newServiceBrokerService(informerFactory.Servicecatalog().V1beta1().ServiceBrokers().Informer())

	bindingService := newServiceBindingService(client.ServicecatalogV1beta1(), informerFactory.Servicecatalog().V1beta1().ServiceBindings().Informer(), name.Generate)

	bindingUsageClient, err := bindingUsageClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Binding Usage Clientset")
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	bindingUsageInformerFactory := bindingUsageInformers.NewSharedInformerFactory(bindingUsageClient, informerResyncPeriod)
	usageKindService := newUsageKindService(bindingUsageClient.ServicecatalogV1alpha1(), dynamicClient, bindingUsageInformerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer())
	bindingUsageService := newServiceBindingUsageService(bindingUsageClient.ServicecatalogV1alpha1(), bindingUsageInformerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages().Informer(), bindingService, name.Generate)

	return &Container{
		Resolver: &Resolver{
			informerFactory:              informerFactory,
			bindingUsageInformerFactory:  bindingUsageInformerFactory,
			serviceInstanceResolver:      newServiceInstanceResolver(instanceService, clusterServicePlanService, clusterServiceClassService, servicePlanService, serviceClassService),
			clusterServiceClassResolver:  newClusterServiceClassResolver(clusterServiceClassService, clusterServicePlanService, instanceService, asyncApiSpecGetter, apiSpecGetter, contentGetter),
			serviceClassResolver:         newServiceClassResolver(serviceClassService, servicePlanService, instanceService, asyncApiSpecGetter, apiSpecGetter, contentGetter),
			clusterServiceBrokerResolver: newClusterServiceBrokerResolver(clusterServiceBrokerService),
			serviceBrokerResolver:        newServiceBrokerResolver(serviceBrokerService),
			serviceBindingResolver:       newServiceBindingResolver(bindingService),
			serviceBindingUsageResolver:  newServiceBindingUsageResolver(bindingUsageService, bindingService),
			usageKindResolver:            newUsageKindResolver(usageKindService),
			bindableResourcesResolver:    newBindableResourcesResolver(usageKindService),
		},
		ServiceBindingUsageLister: bindingUsageService,
		ServiceBindingGetter:      bindingService,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
	r.bindingUsageInformerFactory.Start(stopCh)
	r.bindingUsageInformerFactory.WaitForCacheSync(stopCh)
}
