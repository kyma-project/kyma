package roles

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var roleGroupVersionResource = schema.GroupVersionResource{
	Version:  v1.SchemeGroupVersion.Version,
	Group:    v1.SchemeGroupVersion.Group,
	Resource: "roles",
}

var clusterRoleGroupVersionResource = schema.GroupVersionResource{
	Version:  v1.SchemeGroupVersion.Version,
	Group:    v1.SchemeGroupVersion.Group,
	Resource: "clusterroles",
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

type Service struct {
	*resource.Service
}

func NewRoleService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {

	return serviceFactory.ForResource(roleGroupVersionResource), nil
}

func NewClusterRoleService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(clusterRoleGroupVersionResource), nil
}

func NewRoleBindingService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(roleBindingGroupVersionResource), nil
}

func NewClusterRoleBindingService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(clusterRoleBindingGroupVersionResource), nil
}

func NewClusterRoleBindingEventHandler(channel chan<- *gqlschema.ClusterRoleBindingEvent, filter func(binding v1.ClusterRoleBinding) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &ClusterRoleBindingEventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1.ClusterRoleBinding{},
		}
	}
}

func NewRoleBindingEventHandler(channel chan<- *gqlschema.RoleBindingEvent, filter func(binding v1.RoleBinding) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &RoleBindingEventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1.RoleBinding{},
		}
	}
}
