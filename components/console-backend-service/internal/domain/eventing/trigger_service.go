package eventing

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate mockery -name=assetSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=assetSvc -case=underscore -output disabled -outpkg disabled
type triggerSvc interface {
	Create(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error)
	CreateMany(triggers []*v1alpha1.Trigger) ([]*v1alpha1.Trigger, error)
	Delete(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error)
	DeleteMany(triggers []*v1alpha1.Trigger) ([]*v1alpha1.Trigger, error)
	List(namespace string, subscriber *gqlschema.SubscriberInput) ([]*v1alpha1.Trigger, error)
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}

type triggerService struct {
	*resource.Service
	notifier notifierResource.Notifier
}

func newTriggerService(serviceFactory *resource.ServiceFactory) triggerSvc {
	notifier := notifierResource.NewNotifier()

	svc := &triggerService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Resource: "trigger",
		}),
		notifier: notifier,
	}

	return svc
}

func (s *triggerService) List(namespace string, subscriber *gqlschema.SubscriberInput) ([]*v1alpha1.Trigger, error) {
	items := make([]*v1alpha1.Trigger, 0)
	err := s.ListInIndex("namespace", namespace, &items)
	if err != nil {
		return nil, err
	}

	if subscriber == nil {
		return items, nil
	}

	var triggers []*v1alpha1.Trigger
	if subscriber.Ref != nil {
		triggers = filterByRef(subscriber.Ref, triggers)
	} else if subscriber.URI != nil {
		triggers = filterByUri(*subscriber.URI, triggers)
	}

	return triggers, nil
}

var triggerTypeMeta = metav1.TypeMeta{
	Kind:       "Trigger",
	APIVersion: "eventing.knative.dev/v1alpha1",
}

func (s *triggerService) Create(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error) {
	trigger.TypeMeta = triggerTypeMeta

	u, err := resource.ToUnstructured(trigger)
	if err != nil {
		return nil, err
	}

	created, err := s.Client.Namespace(trigger.ObjectMeta.Namespace).Create(u, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	var createdTrigger *v1alpha1.Trigger
	err = resource.FromUnstructured(created, createdTrigger)
	if err != nil {
		return nil, err
	}

	return createdTrigger, nil
}

func (s *triggerService) CreateMany(triggers []*v1alpha1.Trigger) ([]*v1alpha1.Trigger, error) {
	var createdTriggers []*v1alpha1.Trigger
	for _, trigger := range triggers {
		created, err := s.Create(trigger)
		if err != nil {
			_, _ = s.DeleteMany(createdTriggers)
			return nil, errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", Trigger, trigger.Name, trigger.Namespace)
		}
		createdTriggers = append(createdTriggers, created)
	}
	return createdTriggers, nil
}

func (s *triggerService) Delete(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error) {

}

func (s *triggerService) DeleteMany(triggers []*v1alpha1.Trigger) ([]*v1alpha1.Trigger, error) {

}

func (s *triggerService) Subscribe(listener notifierResource.Listener) {
	s.notifier.AddListener(listener)
}

func (s *triggerService) Unsubscribe(listener notifierResource.Listener) {
	s.notifier.DeleteListener(listener)
}


func filterByUri(uri string, triggers []*v1alpha1.Trigger) []*v1alpha1.Trigger {
	var items []*v1alpha1.Trigger
	for _, trigger := range triggers {
		if uri == trigger.Spec.Subscriber.URI.Path {
			items = append(items, trigger)
		}
	}
	return items
}

func filterByRef(ref *gqlschema.SubscriberRefInput, triggers []*v1alpha1.Trigger) []*v1alpha1.Trigger {
	var items []*v1alpha1.Trigger
	for _, trigger := range triggers {
		if ref != nil &&
			ref.APIVersion == trigger.Spec.Subscriber.Ref.APIVersion &&
			ref.Kind == trigger.Spec.Subscriber.Ref.Kind &&
			ref.Name == trigger.Spec.Subscriber.Ref.Name &&
			ref.Namespace == trigger.Spec.Subscriber.Ref.Namespace {
			items = append(items, trigger)
		}
	}
	return items
}