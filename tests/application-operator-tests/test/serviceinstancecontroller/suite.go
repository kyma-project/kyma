package serviceinstancecontroller

import (
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	testRealeaseName  = "%s-gateway"
	testNamespaceName = "operator-test-%s"
	testInstanceName  = "operator-test-%s"

	defaultCheckInterval = 2 * time.Second
	waitBeforeCheck      = 2 * time.Second
)

type TestSuite struct {
	serviceInstance string
	namespace       string
	releaseName     string

	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient
	k8sChecker *testkit.K8sResourceChecker

	installationTimeout time.Duration
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	randomString := rand.String(4)

	instance := fmt.Sprintf(testInstanceName, randomString)
	namespace := fmt.Sprintf(testNamespaceName, randomString)
	releaseName := fmt.Sprintf(testRealeaseName, namespace)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(namespace)
	require.NoError(t, err)

	helmClient, err := testkit.NewHelmClient(config.TillerHost, config.TillerTLSKeyFile, config.TillerTLSCertificateFile, config.TillerTLSSkipVerify)
	require.NoError(t, err)

	k8sResourcesChecker := testkit.NewServiceInstanceK8SChecker(k8sResourcesClient, releaseName)

	return &TestSuite{
		serviceInstance:     instance,
		namespace:           namespace,
		releaseName:         releaseName,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		k8sChecker:          k8sResourcesChecker,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeoutSeconds),
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: ts.namespace,
		},
	}
	namespace, err := ts.k8sClient.CreateNamespace(ns)
	require.NoError(t, err)
	require.NotEmpty(t, namespace)
}

func (ts *TestSuite) DeleteTestNamespace(t *testing.T) {
	err := ts.k8sClient.DeleteNamespace()
	require.NoError(t, err)
}

func (ts *TestSuite) CreateServiceInstance(t *testing.T) {
	serviceInstance := &v1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      ts.serviceInstance,
			Namespace: ts.namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: "781757f2-8a04-42c0-9015-62905b0f5431",
				ServicePlanExternalName:  "default",
			},
		},
	}
	instance, err := ts.k8sClient.CreateServiceInstance(serviceInstance)

	require.NoError(t, err)
	require.NotEmpty(t, instance)
}

func (ts *TestSuite) DeleteServiceInstance(t *testing.T) {
	err := ts.k8sClient.DeleteServiceInstance(ts.serviceInstance)
	require.NoError(t, err)
}

func (ts *TestSuite) WaitForReleaseToInstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		return ts.helmClient.IsInstalled(ts.releaseName)
	})
	require.NoError(t, err, "Received timeout while waiting for release to install")
}

func (ts *TestSuite) WaitForReleaseToUninstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, ts.helmReleaseNotExist)
	require.NoError(t, err, "Received timeout while waiting for release to uninstall")
}

func (ts *TestSuite) CheckK8sResourcesDeployed(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.CheckK8sResources(t, ts.k8sChecker.CheckResourceDeployed)
}

func (ts *TestSuite) CheckK8sResourceRemoved(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.CheckK8sResources(t, ts.k8sChecker.CheckResourceRemoved)
}

func (ts *TestSuite) Cleanup() {
	//errors are not handled because resources can be deleted already
	ts.k8sClient.DeleteServiceInstance(ts.releaseName)
	ts.k8sClient.DeleteNamespace()
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	return !ts.helmClient.IsInstalled(ts.releaseName)
}
