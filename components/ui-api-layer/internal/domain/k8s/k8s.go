package k8s

import (
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/informers"
	k8sClientset "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type ApplicationLister interface {
	ListInNamespace(namespace string) ([]*v1alpha1.Application, error)
	ListNamespacesFor(reName string) ([]string, error)
}

type Resolver struct {
	*namespaceResolver
	*secretResolver
	*deploymentResolver
	*resourceQuotaResolver
	*resourceQuotaStatusResolver
	*limitRangeResolver
	*podResolver
	*replicaSetResolver
	informerFactory informers.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, applicationRetriever shared.ApplicationRetriever, scRetriever shared.ServiceCatalogRetriever, scaRetriever shared.ServiceCatalogAddonsRetriever) (*Resolver, error) {
	client, err := v1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Client")
	}

	clientset, err := k8sClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Client")
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, informerResyncPeriod)

	namespaceService := newNamespaceService(client.Namespaces())
	deploymentService, err := newDeploymentService(informerFactory.Apps().V1beta2().Deployments().Informer())
	if err != nil {
		return nil, errors.Wrap(err, "while creating deployment service")
	}
	limitRangeService := newLimitRangeService(informerFactory.Core().V1().LimitRanges().Informer())
	podService := newPodService(informerFactory.Core().V1().Pods().Informer(), client)

	replicaSetService := newReplicaSetService(informerFactory.Apps().V1().ReplicaSets().Informer(), clientset.AppsV1())
	resourceQuotaService := newResourceQuotaService(informerFactory.Core().V1().ResourceQuotas().Informer(),
		informerFactory.Apps().V1().ReplicaSets().Informer(), informerFactory.Apps().V1().StatefulSets().Informer(), client)
	resourceQuotaStatusService := newResourceQuotaStatusService(resourceQuotaService, resourceQuotaService, resourceQuotaService, limitRangeService)

	return &Resolver{
		namespaceResolver:           newNamespaceResolver(namespaceService, applicationRetriever),
		secretResolver:              newSecretResolver(client),
		deploymentResolver:          newDeploymentResolver(deploymentService, scRetriever, scaRetriever),
		podResolver:                 newPodResolver(podService),
		replicaSetResolver:          newReplicaSetResolver(replicaSetService),
		limitRangeResolver:          newLimitRangeResolver(limitRangeService),
		resourceQuotaResolver:       newResourceQuotaResolver(resourceQuotaService),
		resourceQuotaStatusResolver: newResourceQuotaStatusResolver(resourceQuotaStatusService),
		informerFactory:             informerFactory,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
