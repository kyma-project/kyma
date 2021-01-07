package k8sNew

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Service struct {
	*resource.Service
}

var limitRangesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1.SchemeGroupVersion.Version,
	Group:    v1.SchemeGroupVersion.Group,
	Resource: "limitranges",
}

func NewLimitRangesService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(limitRangesGroupVersionResource), nil
}

var resourceQuotasGroupVersionResource = schema.GroupVersionResource{
	Version:  v1.SchemeGroupVersion.Version,
	Group:    v1.SchemeGroupVersion.Group,
	Resource: "resourcequotas",
}

func NewResourceQuotasService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(resourceQuotasGroupVersionResource), nil
}
