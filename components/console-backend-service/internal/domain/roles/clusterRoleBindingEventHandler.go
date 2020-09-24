package roles

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/rbac/v1"
)

type ClusterRoleBindingEventHandler struct {
	channel chan<- *gqlschema.ClusterRoleBindingEvent
	filter  func(binding v1.ClusterRoleBinding) bool
	res     *v1.ClusterRoleBinding
}

func (h *ClusterRoleBindingEventHandler) K8sResource() interface{} {
	return h.res
}

func (h *ClusterRoleBindingEventHandler) ShouldNotify() bool {
	return h.filter(*h.res)
}

func (h *ClusterRoleBindingEventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
	h.channel <- &gqlschema.ClusterRoleBindingEvent{
		Type:               eventType,
		ClusterRoleBinding: h.res,
	}
}
