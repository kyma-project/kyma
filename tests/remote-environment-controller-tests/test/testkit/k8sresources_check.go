package testkit

import (
	"testing"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gatewayNameFormat = "%s-gateway"
	gatewayRoleFormat = "%s-gateway-role"
	gatewayRoleBindingFormat = "%s-gateway-rolebinding"
	gatewayApiFormat = "%s-gateway-external-api"
	gatewayEchoFormat = "%s-gateway-echo"

	eventServiceNameFromat = "%s-event-service"
	eventServiceApiFormat = "%s-gateway-external-api"
)

type K8sChecker interface {
	CheckK8sResources(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{}))
}

func NewK8sResourceChecker(reName string, client K8sResourcesClient) K8sChecker {
	return &k8sChecker{
		remoteEnvName: reName,
		client: client,
	}
}

type k8sChecker struct {
	remoteEnvName string
	client K8sResourcesClient
}

func (checker *k8sChecker) CheckK8sResources(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	checker.checkDeployments(t, errCheckFunc, resourceCheckFunc)
	checker.checkIngress(t, errCheckFunc, resourceCheckFunc)
	checker.checkRole(t, errCheckFunc, resourceCheckFunc)
	checker.checkRoleBinding(t, errCheckFunc, resourceCheckFunc)
	checker.checkServices(t, errCheckFunc, resourceCheckFunc)
}

func (checker *k8sChecker) checkDeployments(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	gatewayName := fmt.Sprintf(gatewayNameFormat, checker.remoteEnvName)
	eventServiceName := fmt.Sprintf(eventServiceNameFromat, checker.remoteEnvName)

	gatewayDeploy, err := checker.client.GetDeployment(gatewayName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, gatewayDeploy)

	eventsDeploy, err := checker.client.GetDeployment(eventServiceName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, eventsDeploy)
}

func (checker *k8sChecker) checkIngress(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	ingressName := fmt.Sprintf(gatewayNameFormat, checker.remoteEnvName)

	ingress, err := checker.client.GetIngress(ingressName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, ingress)
}

func (checker *k8sChecker) checkRole(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	roleName := fmt.Sprintf(gatewayRoleFormat, checker.remoteEnvName)

	role, err := checker.client.GetRole(roleName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, role)
}

func (checker *k8sChecker) checkRoleBinding(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	roleBindingName := fmt.Sprintf(gatewayRoleBindingFormat, checker.remoteEnvName)

	role, err := checker.client.GetRoleBinding(roleBindingName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, role)
}

func (checker *k8sChecker) checkServices(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	gatewayApiName := fmt.Sprintf(gatewayApiFormat, checker.remoteEnvName)
	gatewayEchoName := fmt.Sprintf(gatewayEchoFormat, checker.remoteEnvName)
	eventsApiName := fmt.Sprintf(eventServiceApiFormat, checker.remoteEnvName)

	gatewayApiSvc, err := checker.client.GetService(gatewayApiName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, gatewayApiSvc)

	gatewayEchoSvc, err := checker.client.GetService(gatewayEchoName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, gatewayEchoSvc)

	eventsSvc, err := checker.client.GetService(eventsApiName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, eventsSvc)
}
