package roles

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/rbac/v1"
)

type RoleBindingEventHandler struct {
	channel chan<- *gqlschema.RoleBindingEvent
	filter  func(binding v1.RoleBinding) bool
	res     *v1.RoleBinding
}

func (h *RoleBindingEventHandler) K8sResource() interface{} {
	return h.res
}

func (h *RoleBindingEventHandler) ShouldNotify() bool {
	return h.filter(*h.res)
}

func (h *RoleBindingEventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
	h.channel <- &gqlschema.RoleBindingEvent{
		Type:        eventType,
		RoleBinding: h.res,
	}
}
