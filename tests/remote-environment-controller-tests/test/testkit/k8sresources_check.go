package testkit

import (
	"testing"
	"github.com/stretchr/testify/require"
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

func CheckDeployments(t *testing.T, reName string, client K8sResourcesClient) {
	gatewayName := fmt.Sprintf(gatewayNameFormat, reName)
	eventServiceName := fmt.Sprintf(eventServiceNameFromat, reName)

	gatewayDeploy, err := client.GetDeployment(gatewayName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, gatewayDeploy)

	eventsDeploy, err := client.GetDeployment(eventServiceName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, eventsDeploy)
}

func CheckIngress(t *testing.T, reName string, client K8sResourcesClient) {
	ingressName := fmt.Sprintf(gatewayNameFormat, reName)

	ingress, err := client.GetIngress(ingressName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, ingress)
}

func CheckRole(t *testing.T, reName string, client K8sResourcesClient) {
	roleName := fmt.Sprintf(gatewayRoleFormat, reName)

	role, err := client.GetRole(roleName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, role)
}

func CheckRoleBinding(t *testing.T, reName string, client K8sResourcesClient) {
	roleBindingName := fmt.Sprintf(gatewayRoleBindingFormat, reName)

	role, err := client.GetRole(roleBindingName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, role)
}

func CheckServices(t *testing.T, reName string, client K8sResourcesClient) {
	gatewayApiName := fmt.Sprintf(gatewayApiFormat, reName)
	gatewayEchoName := fmt.Sprintf(gatewayEchoFormat, reName)
	eventsApiName := fmt.Sprintf(eventServiceApiFormat, reName)

	gatewayApiSvc, err := client.GetService(gatewayApiName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, gatewayApiSvc)

	gatewayEchoSvc, err := client.GetService(gatewayEchoName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, gatewayEchoSvc)

	eventsSvc, err := client.GetService(eventsApiName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, eventsSvc)
}
