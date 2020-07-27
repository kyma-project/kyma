package oauth

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/ory/hydra-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var oAuth2ClientKind = "OAuth2Client"
var oAuth2ClientGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "oauth2clients",
}

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(oAuth2ClientGroupVersionResource), nil
}

func NewEventHandler(channel chan<- *gqlschema.OAuth2ClientEvent, filter func(client v1alpha1.OAuth2Client) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &EventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1alpha1.OAuth2Client{},
		}
	}
}

type EventHandler struct {
	channel chan<- *gqlschema.OAuth2ClientEvent
	filter  func(client v1alpha1.OAuth2Client) bool
	res     *v1alpha1.OAuth2Client
}

func (h *EventHandler) K8sResource() interface{} {
	return h.res
}

func (h *EventHandler) ShouldNotify() bool {
	return h.filter(*h.res)
}

func (h *EventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
	h.channel <- &gqlschema.OAuth2ClientEvent{
		Type:   eventType,
		Client: h.res,
	}
}
