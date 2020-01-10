package application

import (
	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	mappingLister "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8sClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func NewApplicationService(cfg Config, aCli dynamic.NamespaceableResourceInterface, mCli mappingCli.ApplicationconnectorV1alpha1Interface, mInformer cache.SharedIndexInformer, mLister mappingLister.ApplicationMappingLister, appInformer cache.SharedIndexInformer) (*applicationService, error) {
	return newApplicationService(cfg, aCli, mCli, mInformer, mLister, appInformer)
}

func NewEventActivationService(informer cache.SharedIndexInformer) *eventActivationService {
	return newEventActivationService(informer)
}

func NewEventActivationResolver(service eventActivationLister, rafterRetriever shared.RafterRetriever) *eventActivationResolver {
	return newEventActivationResolver(service, rafterRetriever)
}

func (r *PluggableContainer) SetFakeClient() {
	r.cfg.mappingClient = mappingClient.NewSimpleClientset()
	scheme := runtime.NewScheme()
	r.cfg.appClient = fake.NewSimpleDynamicClient(scheme)
	r.cfg.k8sCli = k8sClient.NewSimpleClientset()
}
