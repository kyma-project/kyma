package testkit

import (
	"fmt"
	v1apps "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
	"github.com/stretchr/testify/require"
)

const (
	gatewayNameFormat        = "%s-gateway"
	gatewayRoleFormat        = "%s-gateway-role"
	gatewayRoleBindingFormat = "%s-gateway-rolebinding"
	gatewayApiFormat         = "%s-gateway-external-api"

	eventServiceNameFormat = "%s-event-service"
	eventServiceApiFormat  = "%s-gateway-external-api"
)

type K8sChecker interface {
	CheckK8sResources(t *testing.T, reName string)
}

func NewK8sCheckerForCreatedResources(client K8sResourcesClient, retryCount int, retryWaitTime time.Duration) K8sChecker {
	return &k8sChecker{
		client:        client,
		retryCount:    retryCount,
		retryWaitTime: retryWaitTime,
		resourcesShouldExist:true,
		errCheckFunc:requireNoError,
		resourceCheckFunc:requireNotEmpty,
	}
}

func NewK8sCheckerForDeletedResources(client K8sResourcesClient, retryCount int, retryWaitTime time.Duration) K8sChecker {
	return &k8sChecker{
		client:        client,
		retryCount:    retryCount,
		retryWaitTime: retryWaitTime,
		resourcesShouldExist:false,
		errCheckFunc:requireError,
		resourceCheckFunc:requireEmpty,
	}
}

type k8sChecker struct {
	client        K8sResourcesClient
	retryCount    int
	retryWaitTime time.Duration
	resourcesShouldExist bool
	errCheckFunc  func(*testing.T, error)
	resourceCheckFunc func(*testing.T, interface{})
}

func (checker *k8sChecker) CheckK8sResources(t *testing.T, reName string) {
	checker.checkDeployments(t, reName)
	checker.checkIngress(t, reName)
	checker.checkRole(t, reName)
	checker.checkRoleBinding(t, reName)
	checker.checkServices(t, reName)
}

func (checker *k8sChecker) checkDeployments(t *testing.T, reName string) {
	gatewayName := fmt.Sprintf(gatewayNameFormat, reName)
	eventServiceName := fmt.Sprintf(eventServiceNameFormat, reName)

	var gatewayDeploy, eventsDeploy *v1apps.Deployment
	var gatewayErr, eventsErr error

	if checker.resourcesShouldExist {
		gatewayDeploy, gatewayErr = checker.client.GetDeployment(gatewayName, v1.GetOptions{})
		eventsDeploy, eventsErr = checker.client.GetDeployment(eventServiceName, v1.GetOptions{})
	} else {
		// Deleting deployments is slow (all the pods needs to be deleted) so that we need to retry checks
		gatewayDeploy, gatewayErr = checker.getDeletedDeployment(gatewayName, v1.GetOptions{})
		eventsDeploy, eventsErr = checker.getDeletedDeployment(eventServiceName, v1.GetOptions{})
	}

	checker.errCheckFunc(t, gatewayErr)
	checker.resourceCheckFunc(t, gatewayDeploy)

	checker.errCheckFunc(t, eventsErr)
	checker.resourceCheckFunc(t, eventsDeploy)
}

func (checker *k8sChecker) checkIngress(t *testing.T, reName string) {
	ingressName := fmt.Sprintf(gatewayNameFormat, reName)

	ingress, err := checker.client.GetIngress(ingressName, v1.GetOptions{})
	checker.errCheckFunc(t, err)
	checker.resourceCheckFunc(t, ingress)
}

func (checker *k8sChecker) checkRole(t *testing.T, reName string) {
	roleName := fmt.Sprintf(gatewayRoleFormat, reName)

	role, err := checker.client.GetRole(roleName, v1.GetOptions{})
	checker.errCheckFunc(t, err)
	checker.resourceCheckFunc(t, role)
}

func (checker *k8sChecker) checkRoleBinding(t *testing.T, reName string) {
	roleBindingName := fmt.Sprintf(gatewayRoleBindingFormat, reName)

	role, err := checker.client.GetRoleBinding(roleBindingName, v1.GetOptions{})
	checker.errCheckFunc(t, err)
	checker.resourceCheckFunc(t, role)
}

func (checker *k8sChecker) checkServices(t *testing.T, reName string) {
	gatewayApiName := fmt.Sprintf(gatewayApiFormat, reName)
	eventsApiName := fmt.Sprintf(eventServiceApiFormat, reName)

	gatewayApiSvc, err := checker.client.GetService(gatewayApiName, v1.GetOptions{})
	checker.errCheckFunc(t, err)
	checker.resourceCheckFunc(t, gatewayApiSvc)

	eventsSvc, err := checker.client.GetService(eventsApiName, v1.GetOptions{})
	checker.errCheckFunc(t, err)
	checker.resourceCheckFunc(t, eventsSvc)
}

func (checker *k8sChecker) getDeletedDeployment(name string, options v1.GetOptions) (*v1apps.Deployment, error) {
	var deployment *v1apps.Deployment
	var err error

	for i := 0; i < checker.retryCount; i++ {
		deployment, err = checker.client.GetDeployment(name, v1.GetOptions{})
		if err != nil && k8serrors.IsNotFound(err) {
			break
		}
		time.Sleep(checker.retryWaitTime)
	}

	return deployment, err
}

func requireError(t *testing.T, err error) {
	require.Error(t, err)
	require.True(t, k8serrors.IsNotFound(err))
}

func requireNoError(t *testing.T, err error) {
	require.NoError(t, err)
}

func requireNotEmpty(t *testing.T, obj interface{}) {
	require.NotEmpty(t, obj)
}

func requireEmpty(t *testing.T, obj interface{}) {
	require.Empty(t, obj)
}
