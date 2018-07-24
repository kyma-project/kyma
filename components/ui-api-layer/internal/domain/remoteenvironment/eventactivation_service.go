package remoteenvironment

import (
	"fmt"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type eventActivationService struct {
	informer cache.SharedIndexInformer
}

func newEventActivationService(informer cache.SharedIndexInformer) *eventActivationService {
	return &eventActivationService{
		informer: informer,
	}
}

func (svc *eventActivationService) List(environment string) ([]*v1alpha1.EventActivation, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", environment)
	if err != nil {
		return nil, err
	}

	return svc.toEventActivations(items)
}

func (svc *eventActivationService) toEventActivation(item interface{}) (*v1alpha1.EventActivation, error) {
	eventActivation, ok := item.(*v1alpha1.EventActivation)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *EventActivation", item)
	}

	return eventActivation, nil
}

func (svc *eventActivationService) toEventActivations(items []interface{}) ([]*v1alpha1.EventActivation, error) {
	var eventActivations []*v1alpha1.EventActivation
	for _, item := range items {
		eventActivation, err := svc.toEventActivation(item)
		if err != nil {
			return nil, err
		}

		eventActivations = append(eventActivations, eventActivation)
	}

	return eventActivations, nil
}
