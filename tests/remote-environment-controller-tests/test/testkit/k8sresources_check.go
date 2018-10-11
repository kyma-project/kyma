package testkit

import (
	"fmt"
	v1apps "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	proxyServiceNameFormat        = "%s-proxy-service"
	proxyServiceRoleFormat        = "%s-proxy-service-role"
	proxyServiceRoleBindingFormat = "%s-proxy-service-rolebinding"
	proxyServiceApiFormat         = "%s-proxy-service-external-api"

	eventServiceNameFormat = "%s-event-service"
	eventServiceApiFormat  = "%s-event-service-external-api"
)

type K8sChecker interface {
	CheckK8sResources(t *testing.T, resourcesShouldExist bool, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{}))
}

func NewK8sResourceChecker(reName string, client K8sResourcesClient, retryCount int, retryWaitTime time.Duration) K8sChecker {
	return &k8sChecker{
		remoteEnvName: reName,
		client:        client,
		retryCount:    retryCount,
		retryWaitTime: retryWaitTime,
	}
}

type k8sChecker struct {
	remoteEnvName string
	client        K8sResourcesClient
	retryCount    int
	retryWaitTime time.Duration
}

func (checker *k8sChecker) CheckK8sResources(t *testing.T, resourcesShouldExist bool, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	checker.checkDeployments(t, resourcesShouldExist, errCheckFunc, resourceCheckFunc)
	checker.checkIngress(t, errCheckFunc, resourceCheckFunc)
	checker.checkRole(t, errCheckFunc, resourceCheckFunc)
	checker.checkRoleBinding(t, errCheckFunc, resourceCheckFunc)
	checker.checkServices(t, errCheckFunc, resourceCheckFunc)
}

func (checker *k8sChecker) checkDeployments(t *testing.T, resourcesShouldExist bool, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	proxyServiceName := fmt.Sprintf(proxyServiceNameFormat, checker.remoteEnvName)
	eventServiceName := fmt.Sprintf(eventServiceNameFormat, checker.remoteEnvName)

	var proxyServiceDeploy, eventsDeploy *v1apps.Deployment
	var proxyServiceErr, eventsErr error

	if resourcesShouldExist {
		proxyServiceDeploy, proxyServiceErr = checker.client.GetDeployment(proxyServiceName, v1.GetOptions{})
		eventsDeploy, eventsErr = checker.client.GetDeployment(eventServiceName, v1.GetOptions{})
	} else {
		// Deleting deployments is slow (all the pods needs to be deleted) so that we need to retry checks
		proxyServiceDeploy, proxyServiceErr = checker.getDeletedDeployment(proxyServiceName, v1.GetOptions{})
		eventsDeploy, eventsErr = checker.getDeletedDeployment(eventServiceName, v1.GetOptions{})
	}

	errCheckFunc(t, proxyServiceErr)
	resourceCheckFunc(t, proxyServiceDeploy)

	errCheckFunc(t, eventsErr)
	resourceCheckFunc(t, eventsDeploy)
}

func (checker *k8sChecker) checkIngress(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	ingressName := fmt.Sprintf(proxyServiceNameFormat, checker.remoteEnvName)

	ingress, err := checker.client.GetIngress(ingressName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, ingress)
}

func (checker *k8sChecker) checkRole(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	roleName := fmt.Sprintf(proxyServiceRoleFormat, checker.remoteEnvName)

	role, err := checker.client.GetRole(roleName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, role)
}

func (checker *k8sChecker) checkRoleBinding(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	roleBindingName := fmt.Sprintf(proxyServiceRoleBindingFormat, checker.remoteEnvName)

	role, err := checker.client.GetRoleBinding(roleBindingName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, role)
}

func (checker *k8sChecker) checkServices(t *testing.T, errCheckFunc func(*testing.T, error), resourceCheckFunc func(*testing.T, interface{})) {
	proxyApiName := fmt.Sprintf(proxyServiceApiFormat, checker.remoteEnvName)
	eventsApiName := fmt.Sprintf(eventServiceApiFormat, checker.remoteEnvName)

	proxyApiSvc, err := checker.client.GetService(proxyApiName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, proxyApiSvc)

	eventsSvc, err := checker.client.GetService(eventsApiName, v1.GetOptions{})
	errCheckFunc(t, err)
	resourceCheckFunc(t, eventsSvc)
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
