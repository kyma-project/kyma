package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Deployment

func NewDeploymentService(informer cache.SharedIndexInformer) *deploymentService {
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
