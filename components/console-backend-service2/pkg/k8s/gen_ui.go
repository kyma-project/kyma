
// This file is generated using gen.go. 

package k8s

import (
	types "github.com/kyma-project/kyma/components/console-backend-service2/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

type UiServices struct {
	BackendModules *resource.Service
}

func NewUiServices(serviceFactory *resource.ServiceFactory) *UiServices {
	return &UiServices{
		BackendModules: serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("backendmodules")),
	}
}