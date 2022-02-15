package application

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8sClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func NewApplicationService(cfg Config, aCli dynamic.NamespaceableResourceInterface, mCli dynamic.NamespaceableResourceInterface, mInformer cache.SharedIndexInformer, appInformer cache.SharedIndexInformer) (*applicationService, error) {
	return newApplicationService(cfg, aCli, mCli, mInformer, appInformer)
}

func NewEventActivationService(informer cache.SharedIndexInformer) *eventActivationService {
	return newEventActivationService(informer)
}

func NewEventActivationResolver(service eventActivationLister, rafterRetriever shared.RafterRetriever) *eventActivationResolver {
	return newEventActivationResolver(service, rafterRetriever)
}

func (r *PluggableContainer) SetFakeClient() {
	scheme := runtime.NewScheme()
	r.cfg.mappingClient = fake.NewSimpleDynamicClient(scheme)
	r.cfg.appClient = fake.NewSimpleDynamicClient(scheme)
	r.cfg.k8sCli = k8sClient.NewSimpleClientset()
}
