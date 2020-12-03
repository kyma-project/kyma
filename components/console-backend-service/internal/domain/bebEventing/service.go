package bebEventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	return serviceFactory.ForResource(subscriptionsGroupVersionResource), nil
}

func NewEventHandler(channel chan<- *gqlschema.SubscriptionEvent, filter func(subscription v1alpha1.Subscription) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &EventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1alpha1.Subscription{},
		}
	}
}

type EventHandler struct {
	channel chan<- *gqlschema.SubscriptionEvent
	filter  func(v1alpha1.Subscription) bool
	res     *v1alpha1.Subscription
}

func (h *EventHandler) K8sResource() interface{} {
	return h.res
}

func (h *EventHandler) ShouldNotify() bool {
	return h.filter(*h.res)
}

func (h *EventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
	h.channel <- &gqlschema.SubscriptionEvent{
		Type:    eventType,
		Subscription: h.res,
	}
}
