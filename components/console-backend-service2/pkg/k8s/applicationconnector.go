package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ApplicationConnectorServices struct {
	ApplicationMappings *resource.Service
	Applications        *resource.Service
}

var applicationConnectorGv = schema.GroupVersion{
	Group:   "applicationconnector.kyma-project.io",
	Version: "v1alpha1",
}
var applicationMappingsGvr = applicationConnectorGv.WithResource("applicationmappings")
var applicationsGvr = applicationConnectorGv.WithResource("applications")

func NewApplicationConnectorServices(serviceFactory *resource.ServiceFactory) *ApplicationConnectorServices {
	return &ApplicationConnectorServices{
		ApplicationMappings: serviceFactory.ForResource(applicationMappingsGvr),
		Applications:        serviceFactory.ForResource(applicationsGvr),
	}
}