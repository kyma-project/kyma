package k8s

import (
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Deployment

func NewDeploymentService(informer cache.SharedIndexInformer) *deploymentService {
	return newDeploymentService(informer)
}

func NewDeploymentResolver(service deploymentLister, serviceBindingUsageLister ServiceBindingUsageLister, serviceBindingGetter ServiceBindingGetter) *deploymentResolver {
	return newDeploymentResolver(service, serviceBindingUsageLister, serviceBindingGetter)
}

// Secret

func NewSecretResolver(secretGetter v1.SecretsGetter) *secretResolver {
	return newSecretResolver(secretGetter)
}

func NewResourceQuotaService(rqInformer cache.SharedIndexInformer, rsInformer cache.SharedIndexInformer, ssInformer cache.SharedIndexInformer, podClient *v1.CoreV1Client) *resourceQuotaService {
	return newResourceQuotaService(rqInformer, rsInformer, ssInformer, podClient)
}
