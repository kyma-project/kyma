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
	testRealeaseName       = "%s-gateway"
	testNamespaceName      = "operator-test-%s"
	firstTestInstanceName  = "operator-test-svc-instance-one-%s"
	secondTestInstanceName = "operator-test-svc-instance-two-%s"

	defaultCheckInterval = 2 * time.Second
)

type TestSuite struct {
	serviceInstanceOne string
	serviceInstanceTwo string
	namespace          string
	releaseName        string

	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient
	k8sChecker *testkit.K8sResourceChecker

	installationTimeout time.Duration

	shouldRun bool
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	randomString := rand.String(4)

	instanceOne := fmt.Sprintf(firstTestInstanceName, randomString)
	instanceTwo := fmt.Sprintf(secondTestInstanceName, randomString)
	namespace := fmt.Sprintf(testNamespaceName, randomString)
	releaseName := fmt.Sprintf(testRealeaseName, namespace)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(namespace)
	require.NoError(t, err)

	helmClient, err := testkit.NewHelmClient(config.TillerHost, config.TillerTLSKeyFile, config.TillerTLSCertificateFile, config.TillerTLSSkipVerify)
	require.NoError(t, err)

	k8sResourcesChecker := testkit.NewServiceInstanceK8SChecker(k8sResourcesClient, releaseName)

	return &TestSuite{
		shouldRun:           config.GatewayOncePerNamespace,
		serviceInstanceOne:  instanceOne,
		serviceInstanceTwo:  instanceTwo,
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

func (ts *TestSuite) CreateFirstServiceInstance(t *testing.T) {
	ts.createServiceInstance(t, ts.serviceInstanceOne)
}

func (ts *TestSuite) CreateSecondServiceInstance(t *testing.T) {
	ts.createServiceInstance(t, ts.serviceInstanceTwo)
}

func (ts *TestSuite) createServiceInstance(t *testing.T, svcInstanceName string) {
	serviceInstance := &v1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      svcInstanceName,
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

func (ts *TestSuite) DeleteFirstServiceInstance(t *testing.T) {
	err := ts.k8sClient.DeleteServiceInstance(ts.serviceInstanceOne)
	require.NoError(t, err)
}

func (ts *TestSuite) DeleteSecondServiceInstance(t *testing.T) {
	err := ts.k8sClient.DeleteServiceInstance(ts.serviceInstanceTwo)
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
	ts.k8sChecker.CheckK8sResourcesDeployed(t)
}

func (ts *TestSuite) CheckK8sResourceRemoved(t *testing.T) {
	ts.k8sChecker.CheckK8sResourceRemoved(t)
}

func (ts *TestSuite) Cleanup() {
	//errors are not handled because resources can be deleted already
	ts.k8sClient.DeleteServiceInstance(ts.serviceInstanceOne)
	ts.k8sClient.DeleteServiceInstance(ts.serviceInstanceTwo)
	ts.k8sClient.DeleteNamespace()
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	return !ts.helmClient.IsInstalled(ts.releaseName)
}

func (ts *TestSuite) TestShouldRun() bool {
	return ts.shouldRun
}
