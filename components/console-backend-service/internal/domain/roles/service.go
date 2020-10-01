package roles

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Service struct {
	*resource.Service
}

var roleBindingGroupVersionResource = schema.GroupVersionResource{
	Version:  v1.SchemeGroupVersion.Version,
	Group:    v1.SchemeGroupVersion.Group,
	Resource: "rolebindings",
}

var clusterRoleBindingGroupVersionResource = schema.GroupVersionResource{
	Version:  v1.SchemeGroupVersion.Version,
	Group:    v1.SchemeGroupVersion.Group,
	Resource: "clusterrolebindings",
}
func NewClusterRoleBindingEventHandler(channel chan<- *gqlschema.ClusterRoleBindingEvent, filter func(binding v1.ClusterRoleBinding) bool) resource.EventHandlerProviderXX {
	return func() resource.EventHandlerXX {
		return &ClusterRoleBindingEventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1.ClusterRoleBinding{},
		}
	}
}

//func NewRoleBindingEventHandler(channel chan<- *gqlschema.RoleBindingEvent, filter func(binding v1.RoleBinding) bool) resource.EventHandlerProvider {
//	return func() resource.EventHandler {
//		return &RoleBindingEventHandler{
//			channel: channel,
//			filter:  filter,
//			res:     &v1.RoleBinding{},
//		}
//	}
//}
