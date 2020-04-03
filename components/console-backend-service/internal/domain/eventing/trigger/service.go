package trigger

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"k8s.io/client-go/tools/cache"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	Namespace                      = "namespace"
	Namespace_Subscriber_Uri_Index = "namespace/uri"
	Namespace_Subscriber_Ref_Index = "namespace/apiVersion/kind/subName/subNamespace"
)

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=Service -case=underscore -output disabled -outpkg disabled
type Service interface {
	Create(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error)
	CreateMany(triggers []*v1alpha1.Trigger) ([]*v1alpha1.Trigger, error)
	Delete(trigger gqlschema.TriggerMetadataInput) error
	DeleteMany(triggers []gqlschema.TriggerMetadataInput) error
	List(namespace string, subscriber *gqlschema.SubscriberInput) ([]*v1alpha1.Trigger, error)
	CompareSubscribers(fir *gqlschema.SubscriberInput, sec *v1alpha1.Trigger) bool
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}
type triggerService struct {
	*resource.Service
	notifier  notifierResource.Notifier
	extractor extractor.TriggerUnstructuredExtractor
}

func NewService(serviceFactory *resource.ServiceFactory, extractor extractor.TriggerUnstructuredExtractor) (*triggerService, error) {
	svc := &triggerService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Resource: "triggers",
		}),
		extractor: extractor,
	}

	err := svc.AddIndexers(cache.Indexers{
		Namespace_Subscriber_Uri_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil || entity == nil {
				return nil, errors.New("Cannot convert item")
			} else if entity.Spec.Subscriber.URI == nil {
				return nil, nil
			}

			return []string{fmt.Sprintf("%s/%s", entity.Namespace, entity.Spec.Subscriber.URI.Path)}, nil
		},
		Namespace_Subscriber_Ref_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil || entity == nil {
				return nil, errors.New("Cannot convert item")
			} else if entity.Spec.Subscriber.Ref == nil {
				return nil, nil
			}

			return []string{fmt.Sprintf("%s/%s/%s/%s/%s", entity.Namespace,
				entity.Spec.Subscriber.Ref.APIVersion,
				entity.Spec.Subscriber.Ref.Kind,
				entity.Spec.Subscriber.Ref.Name,
				entity.Spec.Subscriber.Ref.Namespace)}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := notifierResource.NewNotifier()
	svc.Informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (s *triggerService) List(namespace string, subscriber *gqlschema.SubscriberInput) ([]*v1alpha1.Trigger, error) {
	items := make([]*v1alpha1.Trigger, 0)
	if subscriber == nil {
		err := s.ListInIndex(Namespace, namespace, &items)
		if err != nil {
			return nil, err
		}
	} else if subscriber.Ref != nil {
		err := s.ListInIndex(Namespace_Subscriber_Ref_Index,
			fmt.Sprintf("%s/%s/%s/%s/%s",
				namespace,
				subscriber.Ref.APIVersion,
				subscriber.Ref.Kind,
				subscriber.Ref.Name,
				subscriber.Ref.Namespace), &items)
		if err != nil {
			return nil, err
		}
	} else if subscriber.URI != nil {
		err := s.ListInIndex(Namespace_Subscriber_Uri_Index,
			fmt.Sprintf("%s/%s", namespace, *subscriber.URI), &items)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("subscription is not null but it is empty")
	}

	return items, nil
}

var triggerTypeMeta = metav1.TypeMeta{
	Kind:       "Trigger",
	APIVersion: "eventing.knative.dev/v1alpha1",
}

func (s *triggerService) Create(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error) {
	if trigger == nil {
		return nil, errors.New("trigger can't be nil")
	}
	trigger.TypeMeta = triggerTypeMeta

	u, err := resource.ToUnstructured(trigger)
	if err != nil {
		return nil, err
	}

	created, err := s.Client.Namespace(trigger.ObjectMeta.Namespace).Create(u, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	createdTrigger := new(v1alpha1.Trigger)
	err = resource.FromUnstructured(created, createdTrigger)
	if err != nil {
		return nil, err
	}

	return createdTrigger, nil
}

func (s *triggerService) CreateMany(triggers []*v1alpha1.Trigger) ([]*v1alpha1.Trigger, error) {
	createdTriggers := make([]*v1alpha1.Trigger, 0)
	for _, trigger := range triggers {
		created, err := s.Create(trigger)
		if err != nil {
			_ = s.deleteCreated(createdTriggers)
			return nil, errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", pretty.Trigger, trigger.Name, trigger.Namespace)
		}
		createdTriggers = append(createdTriggers, created)
	}
	return createdTriggers, nil
}

func (s *triggerService) deleteCreated(triggers []*v1alpha1.Trigger) error {
	var data []gqlschema.TriggerMetadataInput
	for _, trigger := range triggers {
		data = append(data, gqlschema.TriggerMetadataInput{
			Name:      trigger.Name,
			Namespace: trigger.Namespace,
		})
	}
	return s.DeleteMany(data)
}

func (s *triggerService) Delete(metadata gqlschema.TriggerMetadataInput) error {
	return s.Client.Namespace(metadata.Namespace).Delete(metadata.Name, &metav1.DeleteOptions{})
}

func (s *triggerService) DeleteMany(metadatas []gqlschema.TriggerMetadataInput) error {
	for _, metadata := range metadatas {
		err := s.Delete(metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *triggerService) Subscribe(listener notifierResource.Listener) {
	s.notifier.AddListener(listener)
}

func (s *triggerService) Unsubscribe(listener notifierResource.Listener) {
	s.notifier.DeleteListener(listener)
}

func (s *triggerService) CompareSubscribers(fir *gqlschema.SubscriberInput, sec *v1alpha1.Trigger) bool {
	if fir == nil || sec == nil {
		return false
	}
	if fir.Ref != nil && sec.Spec.Subscriber.URI != nil && fir.URI != nil &&
		sec.Spec.Subscriber.URI.Path == *fir.URI {
		return true
	}
	if s.compareServiceRefs(fir.Ref, sec.Spec.Subscriber.Ref) {
		return true
	}
	return false
}

func (s *triggerService) compareServiceRefs(fir *gqlschema.SubscriberRefInput, sec *duckv1.KReference) bool {
	if fir == nil || sec == nil {
		return false
	}
	return fir.APIVersion == sec.APIVersion && fir.Kind == sec.Kind && fir.Name == sec.Name && fir.Namespace == sec.Namespace
}
