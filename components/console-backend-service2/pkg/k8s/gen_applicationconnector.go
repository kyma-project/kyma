// This file is generated using gen.go.

package k8s

import (
	types "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

type ApplicationConnectorServices struct {
	Applications        *resource.Service
	ApplicationMappings *resource.Service
}

func NewApplicationConnectorServices(serviceFactory *resource.ServiceFactory) *ApplicationConnectorServices {
	return &ApplicationConnectorServices{
		Applications:        serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("applications")),
		ApplicationMappings: serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("applicationmappings")),
	}
}
