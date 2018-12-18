package testkit

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressNameFormat             = "%s-remote-environment"
	proxyServiceDeploymentFormat  = "%s-proxy-service"
	proxyServiceRoleFormat        = "%s-proxy-service-role"
	proxyServiceRoleBindingFormat = "%s-proxy-service-rolebinding"
	proxyServiceSvcFormat         = "%s-proxy-service-external-api"
	eventServiceDeploymentFormat  = "%s-event-service"
	eventServiceSvcFormat         = "%s-event-service-external-api"
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
	reName    string

	resources []k8sResource
}

func NewK8sChecker(client K8sResourcesClient, reName string) *K8sResourceChecker {
	resources := []k8sResource{
		newResource(fmt.Sprintf(ingressNameFormat, reName), "ingress", client.GetIngress),
		newResource(fmt.Sprintf(proxyServiceDeploymentFormat, reName), "deployment", client.GetDeployment),
		newResource(fmt.Sprintf(proxyServiceRoleFormat, reName), "role", client.GetRole),
		newResource(fmt.Sprintf(proxyServiceRoleBindingFormat, reName), "ingress", client.GetRoleBinding),
		newResource(fmt.Sprintf(proxyServiceSvcFormat, reName), "ingress", client.GetService),
		newResource(fmt.Sprintf(eventServiceDeploymentFormat, reName), "ingress", client.GetDeployment),
		newResource(fmt.Sprintf(eventServiceSvcFormat, reName), "ingress", client.GetService),
	}

	return &K8sResourceChecker{
		k8sClient: client,
		reName:    reName,
		resources: resources,
	}
}

func (c *K8sResourceChecker) checkK8sResources(checkFunc func(resource interface{}, err error, failMessage string)) {
	for _, r := range c.resources {
		failMessage := fmt.Sprintf("%s resource %s not handled properly", r.kind, r.name)
		resource, err := r.getFunction(r.name, v1.GetOptions{})
		checkFunc(resource, err, failMessage)
	}
}
