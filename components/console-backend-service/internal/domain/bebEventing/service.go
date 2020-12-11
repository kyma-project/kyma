package bebEventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

var subscriptionsKind = "Subscription"
var subscriptionsGroupVersionResource = schema.GroupVersionResource{
	Version:  "v1alpha1",
	Group:    "eventing.kyma-project.io",
	Resource: "subscriptions",
}

var eventSubscriptionServiceIndex = "service"

func eventSubscriptionServiceIndexKey(namespace string, serviceName string) string {
	return namespace + ":" + serviceName
}

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	service := serviceFactory.ForResource(subscriptionsGroupVersionResource)
	err := service.AddIndexers(cache.Indexers{
		eventSubscriptionServiceIndex: func(obj interface{}) ([]string, error) {
			subscription := &v1alpha1.Subscription{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), subscription)
			if err != nil {
				return nil, err
			}
			if len(subscription.ObjectMeta.OwnerReferences) == 0 {
				return nil, nil
			}
			return []string{eventSubscriptionServiceIndexKey(subscription.ObjectMeta.Namespace, subscription.ObjectMeta.OwnerReferences[0].Name)}, nil
		},
	})
	return service, err
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
		Type:         eventType,
		Subscription: h.res,
	}
}
