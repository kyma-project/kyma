package testkit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	virtualSvcNameFormat                       = "%s-validator"
	applicationGatewayDeploymentFormat         = "%s-application-gateway"
	applicationGatewayRoleFormat               = "%s-application-gateway"
	applicationGatewayRoleBindingFormat        = "%s-application-gateway"
	applicationGatewayClusterRoleFormat        = "%s-application-gateway"
	applicationGatewayClusterRoleBindingFormat = "%s-application-gateway"
	applicationGatewaySvcFormat                = "%s-application-gateway"
	applicationGatewayServiceAccountFormat     = "%s-application-gateway"
	eventServiceDeploymentFormat               = "%s-event-service"
	eventServiceSvcFormat                      = "%s-event-service"
	connectivityValidatorDeploymentFormat      = "%s-connectivity-validator"
	connectivityValidatorSvcFormat             = "%s-validator"

	resourceCheckInterval = 1 * time.Second
	resourceCheckTimeout  = 10 * time.Second
)

type k8sResource struct {
	name        string
	kind        string
	getFunction func(string, v1.GetOptions) (interface{}, error)
}

func newResource(name string, kind string, getFunc func(string, v1.GetOptions) (interface{}, error)) k8sResource {
	return k8sResource{
		name:        name,
		kind:        kind,
		getFunction: getFunc,
	}
}

type K8sResourceChecker struct {
	k8sClient K8sResourcesClient
	appName   string

	resources []k8sResource
}

func NewK8sChecker(client K8sResourcesClient, appName string) *K8sResourceChecker {
	resources := []k8sResource{
		newResource(fmt.Sprintf(virtualSvcNameFormat, appName), "virtualservice", client.GetVirtualService),
		newResource(fmt.Sprintf(applicationGatewayDeploymentFormat, appName), "deployment", client.GetDeployment),
		newResource(fmt.Sprintf(applicationGatewayRoleFormat, appName), "role", client.GetRole),
		newResource(fmt.Sprintf(applicationGatewayRoleBindingFormat, appName), "rolebinding", client.GetRoleBinding),
		newResource(fmt.Sprintf(applicationGatewayClusterRoleFormat, appName), "clusterrole", client.GetClusterRole),
		newResource(fmt.Sprintf(applicationGatewayClusterRoleBindingFormat, appName), "clusterrolebinding", client.GetClusterRoleBinding),
		newResource(fmt.Sprintf(applicationGatewayServiceAccountFormat, appName), "serviceaccount", client.GetServiceAccount),
		newResource(fmt.Sprintf(applicationGatewaySvcFormat, appName), "service", client.GetService),
		newResource(fmt.Sprintf(eventServiceDeploymentFormat, appName), "deployment", client.GetDeployment),
		newResource(fmt.Sprintf(eventServiceSvcFormat, appName), "service", client.GetService),
		newResource(fmt.Sprintf(connectivityValidatorDeploymentFormat, appName), "deployment", client.GetDeployment),
		newResource(fmt.Sprintf(connectivityValidatorSvcFormat, appName), "service", client.GetService),
	}

	return &K8sResourceChecker{
		k8sClient: client,
		appName:   appName,
		resources: resources,
	}
}

func (c *K8sResourceChecker) CheckK8sResources(t *testing.T, checkFunc func(resource interface{}, err error) bool) {
	for _, r := range c.resources {
		failMessage := fmt.Sprintf("%s resource %s not handled properly", r.kind, r.name)

		err := WaitForFunction(resourceCheckInterval, resourceCheckTimeout, func() bool {
			resource, err := r.getFunction(r.name, v1.GetOptions{})
			return checkFunc(resource, err)
		})

		require.NoError(t, err, failMessage)
	}
}
