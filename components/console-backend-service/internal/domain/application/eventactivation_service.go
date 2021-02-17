package application

import (
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	res "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
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

func (svc *eventActivationService) List(namespace string) ([]*v1alpha1.EventActivation, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	return svc.toEventActivations(items)
}

func (svc *eventActivationService) toEventActivation(item interface{}) (*v1alpha1.EventActivation, error) {
	unstructured, err := res.ToUnstructured(item)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting item to unstructured")
	}
	eventActivation := &v1alpha1.EventActivation{}
	err = res.FromUnstructured(unstructured, eventActivation)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventActivation from unstructured")
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
