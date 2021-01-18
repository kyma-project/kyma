package testkit

import (
	"context"
	"fmt"
	"testing"
	"time"

	api "k8s.io/api/apps/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

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
	resourceCheckTimeout  = 20 * time.Second
	waitBeforeCheck       = 2 * time.Second
)

type k8sResource struct {
	context     context.Context
	name        string
	kind        string
	getFunction func(context.Context, string, v1.GetOptions) (interface{}, error)
}

func newResource(ctx context.Context, name string, kind string, getFunc func(context.Context, string, v1.GetOptions) (interface{}, error)) k8sResource {
	return k8sResource{
		context:     ctx,
		name:        name,
		kind:        kind,
		getFunction: getFunc,
	}
}

type K8sResourceChecker struct {
	k8sClient    K8sResourcesClient
	resourceName string
	resources    []k8sResource
}

func NewServiceInstanceK8SChecker(client K8sResourcesClient, releaseName string) *K8sResourceChecker {
	resources := []k8sResource{
		newResource(context.Background(), releaseName, "deployment", cast(client.GetDeployment)),
		newResource(context.Background(), releaseName, "role", client.GetRole),
		newResource(context.Background(), releaseName, "rolebinding", client.GetRoleBinding),
		newResource(context.Background(), releaseName, "serviceaccount", client.GetServiceAccount),
		newResource(context.Background(), releaseName, "service", client.GetService),
	}

	return &K8sResourceChecker{
		k8sClient:    client,
		resourceName: releaseName,
		resources:    resources,
	}
}

func NewAppK8sChecker(client K8sResourcesClient, appName string, checkGateway bool) *K8sResourceChecker {
	ctxBackground := context.Background()

	resources := []k8sResource{
		newResource(ctxBackground, fmt.Sprintf(virtualSvcNameFormat, appName), "virtualservice", client.GetVirtualService),
		newResource(ctxBackground, fmt.Sprintf(eventServiceDeploymentFormat, appName), "deployment", cast(client.GetDeployment)),
		newResource(ctxBackground, fmt.Sprintf(eventServiceSvcFormat, appName), "service", client.GetService),
		newResource(ctxBackground, fmt.Sprintf(connectivityValidatorDeploymentFormat, appName), "deployment", cast(client.GetDeployment)),
		newResource(ctxBackground, fmt.Sprintf(connectivityValidatorSvcFormat, appName), "service", client.GetService),
	}

	if checkGateway {
		resources = append(resources,
			newResource(ctxBackground, fmt.Sprintf(applicationGatewayDeploymentFormat, appName), "deployment", cast(client.GetDeployment)),
			newResource(ctxBackground, fmt.Sprintf(applicationGatewaySvcFormat, appName), "service", client.GetService),
			newResource(ctxBackground, fmt.Sprintf(applicationGatewayRoleFormat, appName), "role", client.GetRole),
			newResource(ctxBackground, fmt.Sprintf(applicationGatewayRoleBindingFormat, appName), "rolebinding", client.GetRoleBinding),
			newResource(ctxBackground, fmt.Sprintf(applicationGatewayClusterRoleFormat, appName), "clusterrole", client.GetClusterRole),
			newResource(ctxBackground, fmt.Sprintf(applicationGatewayClusterRoleBindingFormat, appName), "clusterrolebinding", client.GetClusterRoleBinding),
			newResource(ctxBackground, fmt.Sprintf(applicationGatewayServiceAccountFormat, appName), "serviceaccount", client.GetServiceAccount))
	}

	return &K8sResourceChecker{
		k8sClient:    client,
		resourceName: appName,
		resources:    resources,
	}
}

func (c *K8sResourceChecker) CheckK8sResourcesDeployed(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	c.checkK8sResources(t, c.checkResourceDeployed)
}

func (c *K8sResourceChecker) CheckK8sResourceRemoved(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	c.checkK8sResources(t, c.checkResourceRemoved)
}

func (c *K8sResourceChecker) checkK8sResources(t *testing.T, checkFunc func(resource interface{}, err error) bool) {
	for _, r := range c.resources {
		failMessage := fmt.Sprintf("%s resource %s not handled properly", r.kind, r.name)

		err := WaitForFunction(resourceCheckInterval, resourceCheckTimeout, func() bool {
			resource, err := r.getFunction(context.Background(), r.name, v1.GetOptions{})
			return checkFunc(resource, err)
		})

		require.NoError(t, err, failMessage)
	}
}

func (c *K8sResourceChecker) checkResourceDeployed(_ interface{}, err error) bool {
	if err != nil {
		return false
	}

	return true
}

func (c *K8sResourceChecker) checkResourceRemoved(_ interface{}, err error) bool {
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return true
		}
	}

	return false
}

func cast(deployment func(ctx context.Context, name string, options v1.GetOptions) (*api.Deployment, error)) func(context.Context, string, v1.GetOptions) (interface{}, error) {
	return func(ctx context.Context, name string, options v1.GetOptions) (interface{}, error) {
		return deployment(ctx, name, options)
	}
}
