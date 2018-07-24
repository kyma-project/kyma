package k8s

import (
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/informers"
	k8sClientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type RemoteEnvironmentLister interface {
	ListInEnvironment(environment string) ([]*v1alpha1.RemoteEnvironment, error)
	ListNamespacesFor(reName string) ([]string, error)
}

type Resolver struct {
	*environmentResolver
	*secretResolver
	*deploymentResolver
	*resourceQuotaResolver
	*limitRangeResolver

	informerFactory informers.SharedInformerFactory
}

func New(restConfig *rest.Config, remoteEnvironmentLister RemoteEnvironmentLister, informerResyncPeriod time.Duration, serviceBindingUsageLister ServiceBindingUsageLister, serviceBindingGetter ServiceBindingGetter) (*Resolver, error) {
	client, err := v1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Client")
	}

	clientset, err := k8sClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Client")
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, informerResyncPeriod)

	environmentService := newEnvironmentService(client.Namespaces(), remoteEnvironmentLister)
	deploymentService := newDeploymentService(informerFactory.Apps().V1beta2().Deployments().Informer())
	limitRangeService := newLimitRangeService(informerFactory.Core().V1().LimitRanges().Informer())
	resourceQuotaService := newResourceQuotaService(informerFactory.Core().V1().ResourceQuotas().Informer())

	return &Resolver{
		environmentResolver:   newEnvironmentResolver(environmentService),
		secretResolver:        newSecretResolver(client),
		deploymentResolver:    newDeploymentResolver(deploymentService, serviceBindingUsageLister, serviceBindingGetter),
		limitRangeResolver:    newLimitRangeResolver(limitRangeService),
		resourceQuotaResolver: newResourceQuotaResolver(resourceQuotaService),
		informerFactory:       informerFactory,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
