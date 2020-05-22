package k8s

import (
	types "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

type ApplicationConnectorServices struct {
	ApplicationMappings *resource.Service
	Applications        *resource.Service
}

func NewApplicationConnectorServices(serviceFactory *resource.ServiceFactory) *ApplicationConnectorServices {
	return &ApplicationConnectorServices{
		ApplicationMappings: serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("applicationmappings")),
		Applications:        serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("applications")),
	}
}