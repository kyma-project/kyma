package remoteenvironment

import (
	remoteenvironmentv1alpha1 "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/remoteenvironment/v1alpha1"
	reMappinglister "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/listers/remoteenvironment/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

func NewRemoteEnvironmentService(client remoteenvironmentv1alpha1.RemoteenvironmentV1alpha1Interface, config Config, mappingInformer cache.SharedIndexInformer, mappingLister reMappinglister.EnvironmentMappingLister, reInformer cache.SharedIndexInformer) (*remoteEnvironmentService, error) {
	return newRemoteEnvironmentService(client, config, mappingInformer, mappingLister, reInformer)
}

func NewEventActivationService(informer cache.SharedIndexInformer) *eventActivationService {
	return newEventActivationService(informer)
}

func NewEventActivationResolver(service eventActivationLister, asyncApiSpecGetter AsyncApiSpecGetter) *eventActivationResolver {
	return newEventActivationResolver(service, asyncApiSpecGetter)
}
