package application

import (
	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	mappingLister "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	appClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/fake"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	k8sClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func NewApplicationService(cfg Config, aCli appCli.ApplicationconnectorV1alpha1Interface, mCli mappingCli.ApplicationconnectorV1alpha1Interface, mInformer cache.SharedIndexInformer, mLister mappingLister.ApplicationMappingLister, appInformer cache.SharedIndexInformer) (*applicationService, error) {
	return newApplicationService(cfg, aCli, mCli, mInformer, mLister, appInformer)
}

func NewEventActivationService(informer cache.SharedIndexInformer) *eventActivationService {
	return newEventActivationService(informer)
}

func NewEventActivationResolver(service eventActivationLister, contentRetriever shared.ContentRetriever) *eventActivationResolver {
	return newEventActivationResolver(service, contentRetriever)
}

func (r *PluggableContainer) SetFakeClient() {
	r.cfg.mappingClient = mappingClient.NewSimpleClientset()
	r.cfg.appClient = appClient.NewSimpleClientset()
	r.cfg.k8sCli = k8sClient.NewSimpleClientset()
}
