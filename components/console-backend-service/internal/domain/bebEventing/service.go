package bebEventing

import (
	"fmt"
	"net"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/pkg/apis"
)

var subscriptionsKind = "Subscription"

var subscriptionsGroupVersionResource = schema.GroupVersionResource{
	Version:  "v1alpha1",
	Group:    "eventing.kyma-project.io",
	Resource: "subscriptions",
}


type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(subscriptionsGroupVersionResource)
}

// func NewEventHandler(channel chan<- *gqlschema.TriggerEvent, filter func(trigger v1alpha1.Trigger) bool) resource.EventHandlerProvider {
// 	return func() resource.EventHandler {
// 		return &EventHandler{
// 			channel: channel,
// 			filter:  filter,
// 			res:     &v1alpha1.Trigger{},
// 		}
// 	}
// }

// type EventHandler struct {
// 	channel chan<- *gqlschema.TriggerEvent
// 	filter  func(v1alpha1.Trigger) bool
// 	res     *v1alpha1.Trigger
// }

// func (h *EventHandler) K8sResource() interface{} {
// 	return h.res
// }

// func (h *EventHandler) ShouldNotify() bool {
// 	return h.filter(*h.res)
// }

// func (h *EventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
// 	h.channel <- &gqlschema.TriggerEvent{
// 		Type:    eventType,
// 		Trigger: h.res,
// 	}
// }
