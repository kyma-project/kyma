package resourceskit

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressNameFormat                 = "%s-application"
	applicationProxyDeploymentFormat  = "%s-application-proxy"
	applicationProxyRoleFormat        = "%s-application-proxy-role"
	applicationProxyRoleBindingFormat = "%s-application-proxy-rolebinding"
	applicationProxySvcFormat         = "%s-application-proxy-external-api"
	eventServiceDeploymentFormat      = "%s-event-service"
	eventServiceSvcFormat             = "%s-event-service-external-api"
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

func NewK8sChecker(client K8sResourcesClient) *K8sResourceChecker {
	resources := []k8sResource{
		newResource(fmt.Sprintf(ingressNameFormat, consts.AppName), "ingress", client.GetIngress),
		newResource(fmt.Sprintf(applicationProxyDeploymentFormat, consts.AppName), "deployment", client.GetDeployment),
		newResource(fmt.Sprintf(applicationProxyRoleFormat, consts.AppName), "role", client.GetRole),
		newResource(fmt.Sprintf(applicationProxyRoleBindingFormat, consts.AppName), "ingress", client.GetRoleBinding),
		newResource(fmt.Sprintf(applicationProxySvcFormat, consts.AppName), "ingress", client.GetService),
		newResource(fmt.Sprintf(eventServiceDeploymentFormat, consts.AppName), "ingress", client.GetDeployment),
		newResource(fmt.Sprintf(eventServiceSvcFormat, consts.AppName), "ingress", client.GetService),
	}

	return &K8sResourceChecker{
		k8sClient: client,
		appName:   consts.AppName,
		resources: resources,
	}
}

func (c *K8sResourceChecker) CheckK8sResources() error {
	for _, r := range c.resources {
		_, err := r.getFunction(r.name, v1.GetOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
