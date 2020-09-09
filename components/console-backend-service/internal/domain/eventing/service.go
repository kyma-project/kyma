package eventing

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

var triggersKind = "Trigger"

var triggersGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.SchemeGroupVersion.Version,
	Group:    v1alpha1.SchemeGroupVersion.Group,
	Resource: "triggers",
}

var triggerIndexKey = "ref"

func createTriggerIndexKey(namespace, serviceName string) string {
	return fmt.Sprintf("%s/%s", namespace, serviceName)
}

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	service := serviceFactory.ForResource(triggersGroupVersionResource)
	err := service.AddIndexers(cache.Indexers{
		triggerIndexKey: func(obj interface{}) ([]string, error) {
			trigger := &v1alpha1.Trigger{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), trigger)
			if err != nil {
				return nil, err
			}
			return []string{createTriggerIndexKey(trigger.Namespace, trigger.OwnerReferences[0].Name)}, nil
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
