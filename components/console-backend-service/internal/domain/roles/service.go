package roles

import (
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
