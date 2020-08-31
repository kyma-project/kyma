package roles

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/api/rbac/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//var roleKind = "Role"
var roleGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.SchemeGroupVersion.Version,
	Group:    v1alpha1.SchemeGroupVersion.Group,
	Resource: "roles",
}

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(roleGroupVersionResource), nil
}
