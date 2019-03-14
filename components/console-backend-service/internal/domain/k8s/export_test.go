package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Deployment

func NewDeploymentService(informer cache.SharedIndexInformer) (*deploymentService, error) {
	return newDeploymentService(informer)
}

func NewDeploymentResolver(service deploymentLister, scRetriever shared.ServiceCatalogRetriever, scaRetriever shared.ServiceCatalogAddonsRetriever) *deploymentResolver {
	return newDeploymentResolver(service, scRetriever, scaRetriever)
}

// Secret

func NewSecretResolver(secretGetter v1.SecretsGetter) *secretResolver {
	return newSecretResolver(secretGetter)
}

func NewResourceQuotaService(rqInformer cache.SharedIndexInformer, rsInformer cache.SharedIndexInformer, ssInformer cache.SharedIndexInformer, podClient v1.CoreV1Interface) *resourceQuotaService {
	return newResourceQuotaService(rqInformer, rsInformer, ssInformer, podClient)
}

// Pod

func NewPodResolver(podSvc podSvc) *podResolver {
	return newPodResolver(podSvc)
}

func (r *podResolver) SetInstanceConverter(converter gqlPodConverter) {
	r.podConverter = converter
}

func NewPodService(informer cache.SharedIndexInformer, client v1.CoreV1Interface) *podService {
	return newPodService(informer, client)
}

// ReplicaSet

func NewReplicaSetResolver(replicaSetSvc replicaSetSvc) *replicaSetResolver {
	return newReplicaSetResolver(replicaSetSvc)
}

func (r *replicaSetResolver) SetInstanceConverter(converter gqlReplicaSetConverter) {
	r.replicaSetConverter = converter
}

func NewReplicaSetService(informer cache.SharedIndexInformer, client apps.AppsV1Interface) *replicaSetService {
	return newReplicaSetService(informer, client)
}
