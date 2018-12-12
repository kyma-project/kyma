package application

import (
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	mappingLister "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

func NewApplicationService(cfg Config, aCli appCli.ApplicationconnectorV1alpha1Interface, mCli mappingCli.ApplicationconnectorV1alpha1Interface, mInformer cache.SharedIndexInformer, mLister mappingLister.ApplicationMappingLister, appInformer cache.SharedIndexInformer) (*applicationService, error) {
	return newApplicationService(cfg, aCli, mCli, mInformer, mLister, appInformer)
}

func NewEventActivationService(informer cache.SharedIndexInformer) *eventActivationService {
	return newEventActivationService(informer)
}

func NewEventActivationResolver(service eventActivationLister, asyncApiSpecGetter AsyncApiSpecGetter) *eventActivationResolver {
	return newEventActivationResolver(service, asyncApiSpecGetter)
}
