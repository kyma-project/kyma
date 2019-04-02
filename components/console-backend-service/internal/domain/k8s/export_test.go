package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"k8s.io/client-go/discovery"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

func NewSecretResolver(svc secretSvc) *secretResolver {
	return newSecretResolver(svc)
}

func NewSecretService(informer cache.SharedIndexInformer, client v1.CoreV1Interface) *secretService {
	return newSecretService(informer, client)
}

func (r *secretResolver) SetSecretConverter(converter gqlSecretConverter) {
	r.converter = converter
}

func NewResourceQuotaService(rqInformer cache.SharedIndexInformer, rsInformer cache.SharedIndexInformer, ssInformer cache.SharedIndexInformer, podClient v1.CoreV1Interface) *resourceQuotaService {
	return newResourceQuotaService(rqInformer, rsInformer, ssInformer, podClient)
}

// Pod

func NewPodResolver(podSvc podSvc) *podResolver {
	return newPodResolver(podSvc)
}

func (r *podResolver) SetPodConverter(converter gqlPodConverter) {
	r.podConverter = converter
}

func NewPodService(informer cache.SharedIndexInformer, client v1.CoreV1Interface) *podService {
	return newPodService(informer, client)
}

// Resource

func NewResourceResolver(resourceSvc resourceSvc) *resourceResolver {
	return newResourceResolver(resourceSvc)
}

func (r *resourceResolver) SetResourceConverter(converter gqlResourceConverter) {
	r.gqlResourceConverter = converter
}

func NewResourceService(client discovery.DiscoveryInterface) *resourceService {
	return newResourceService(client)
}

func NewResourceConverter() *resourceConverter {
	return &resourceConverter{}
}

// Service

func NewServiceService(informer cache.SharedIndexInformer, client v1.CoreV1Interface) *serviceService {
	return newServiceService(informer, client)
}

func NewServiceResolver(svc serviceSvc) *serviceResolver {
	return newServiceResolver(svc)
}

func (r *serviceResolver) SetInstanceConverter(converter gqlServiceConverter) {
	r.gqlServiceConverter = converter
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

// ConfigMap

func NewConfigMapResolver(configMapSvc configMapSvc) *configMapResolver {
	return newConfigMapResolver(configMapSvc)
}

func (r *configMapResolver) SetConfigMapConverter(converter gqlConfigMapConverter) {
	r.configMapConverter = converter
}

func NewConfigMapService(informer cache.SharedIndexInformer, client v1.CoreV1Interface) *configMapService {
	return newConfigMapService(informer, client)
}
