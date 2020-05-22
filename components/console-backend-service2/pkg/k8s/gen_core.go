
// This file is generated using gen.go. 

package k8s

import (
	types "k8s.io/api/core/v1"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

type CoreServices struct {
	Namespaces *resource.Service
	Pods *resource.Service
}

func NewCoreServices(serviceFactory *resource.ServiceFactory) *CoreServices {
	return &CoreServices{
		Namespaces: serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("namespaces")),
		Pods: serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("pods")),
	}
}