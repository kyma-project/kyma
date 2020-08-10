package eventing

import (
	"fmt"

	"knative.dev/pkg/apis"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	v1 "knative.dev/pkg/apis/duck/v1"
)

var triggersKind = "Trigger"

var triggersGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.SchemeGroupVersion.Version,
	Group:    v1alpha1.SchemeGroupVersion.Group,
	Resource: "triggers",
}

var triggersSubscriberRefIndex = "subscriberRef"

func createTriggersSubscriberRefIndexKey(namespace string, subscriberRef *v1.KReference) string {
	if subscriberRef == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		namespace,
		subscriberRef.APIVersion,
		subscriberRef.Kind,
		subscriberRef.Name,
		subscriberRef.Namespace)
}

var triggersSubscriberURIIndex = "subscriberURI"

func createTriggersSubscriberRefURIKey(namespace string, uri *apis.URL) string {
	return fmt.Sprintf("%s/%s",
		namespace,
		uri.String())
}

var triggersSubscriberRefUriIndex = "subscriberRefURI"

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	service := serviceFactory.ForResource(triggersGroupVersionResource)
	err := service.AddIndexers(cache.Indexers{
		triggersSubscriberRefIndex: func(obj interface{}) ([]string, error) {
			trigger := &v1alpha1.Trigger{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), trigger)
			if err != nil {
				return nil, err
			}
			return []string{createTriggersSubscriberRefIndexKey(trigger.Namespace, trigger.Spec.Subscriber.Ref)}, nil
		},
		triggersSubscriberURIIndex: func(obj interface{}) ([]string, error) {
			trigger := &v1alpha1.Trigger{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), trigger)
			if err != nil {
				return nil, err
			}
			return []string{createTriggersSubscriberRefURIKey(trigger.Namespace, trigger.Spec.Subscriber.URI)}, nil
		},
	})
	return service, err
}

func NewEventHandler(channel chan<- *gqlschema.TriggerEvent, filter func(trigger v1alpha1.Trigger) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &EventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1alpha1.Trigger{},
		}
	}
}

type EventHandler struct {
	channel chan<- *gqlschema.TriggerEvent
	filter  func(v1alpha1.Trigger) bool
	res     *v1alpha1.Trigger
}

func (h *EventHandler) K8sResource() interface{} {
	return h.res
}

func (h *EventHandler) ShouldNotify() bool {
	return h.filter(*h.res)
}

func (h *EventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
	h.channel <- &gqlschema.TriggerEvent{
		Type:    eventType,
		Trigger: h.res,
	}
}
