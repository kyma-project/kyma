package k8s

import (
	"time"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
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
	*resourceResolver
	*namespaceResolver
	*secretResolver
	*deploymentResolver
	*resourceQuotaResolver
	*resourceQuotaStatusResolver
	*limitRangeResolver
	*podResolver
	*serviceResolver
	*replicaSetResolver
	*configMapResolver
	*selfSubjectRulesResolver
	*versionInfoResolver
	informerFactory informers.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, applicationRetriever shared.ApplicationRetriever, scRetriever shared.ServiceCatalogRetriever, scaRetriever shared.ServiceCatalogAddonsRetriever, systemNamespaces []string) (*Resolver, error) {
	client, err := v1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Client")
	}

	clientset, err := k8sClientset.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Clientset")
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, informerResyncPeriod)

	podService := newPodService(informerFactory.Core().V1().Pods().Informer(), client)
	namespaceSvc, err := newNamespaceService(informerFactory.Core().V1().Namespaces().Informer(), podService, client)
	if err != nil {
		return nil, errors.Wrap(err, "while creating namespace service")
	}

	deploymentService, err := newDeploymentService(informerFactory.Apps().V1().Deployments().Informer())
	if err != nil {
		return nil, errors.Wrap(err, "while creating deployment service")
	}

	limitRangeService := newLimitRangeService(informerFactory.Core().V1().LimitRanges().Informer(), clientset.CoreV1())

	resourceService := newResourceService(clientset.Discovery())
	secretService := newSecretService(informerFactory.Core().V1().Secrets().Informer(), client)

	replicaSetService := newReplicaSetService(informerFactory.Apps().V1().ReplicaSets().Informer(), clientset.AppsV1())
	resourceQuotaService := newResourceQuotaService(informerFactory.Core().V1().ResourceQuotas().Informer(),
		informerFactory.Apps().V1().ReplicaSets().Informer(), informerFactory.Apps().V1().StatefulSets().Informer(), client)
	resourceQuotaStatusService := newResourceQuotaStatusService(resourceQuotaService, resourceQuotaService, resourceQuotaService, limitRangeService)
	configMapService := newConfigMapService(informerFactory.Core().V1().ConfigMaps().Informer(), clientset.CoreV1())
	serviceSvc := newServiceService(informerFactory.Core().V1().Services().Informer(), client)
	selfSubjectRulesService := newSelfSubjectRulesService(clientset.AuthorizationV1())
	return &Resolver{
		resourceResolver:            newResourceResolver(resourceService),
		namespaceResolver:           newNamespaceResolver(namespaceSvc, applicationRetriever, systemNamespaces, podService),
		secretResolver:              newSecretResolver(*secretService),
		deploymentResolver:          newDeploymentResolver(deploymentService, scRetriever, scaRetriever),
		podResolver:                 newPodResolver(podService),
		serviceResolver:             newServiceResolver(serviceSvc),
		replicaSetResolver:          newReplicaSetResolver(replicaSetService),
		limitRangeResolver:          newLimitRangeResolver(limitRangeService),
		resourceQuotaResolver:       newResourceQuotaResolver(resourceQuotaService),
		resourceQuotaStatusResolver: newResourceQuotaStatusResolver(resourceQuotaStatusService),
		configMapResolver:           newConfigMapResolver(configMapService),
		selfSubjectRulesResolver:    newSelfSubjectRulesResolver(selfSubjectRulesService),
		versionInfoResolver:         newVersionInfoResolver(deploymentService),
		informerFactory:             informerFactory,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
