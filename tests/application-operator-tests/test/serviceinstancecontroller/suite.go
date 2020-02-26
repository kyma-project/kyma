package serviceinstancecontroller

import (
	"fmt"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
	"time"
)

const (
	testSIName    = "operator-test-%s"
	testNamespace = "operator-test-ns-%s"

	defaultCheckInterval = 2 * time.Second
	waitBeforeCheck      = 2 * time.Second
)

type TestSuite struct {
	serviceInstance string

	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient
	k8sChecker *testkit.K8sResourceChecker

	installationTimeout time.Duration
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	serviceInstance := fmt.Sprintf(testSIName, rand.String(4))

	namespace := fmt.Sprintf(testNamespace, rand.String(4))

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(namespace)
	require.NoError(t, err)

	helmClient, err := testkit.NewHelmClient(config.TillerHost, config.TillerTLSKeyFile, config.TillerTLSCertificateFile, config.TillerTLSSkipVerify)
	require.NoError(t, err)

	k8sResourcesChecker := testkit.NewServiceInstanceK8SChecker(k8sResourcesClient, serviceInstance)

	return &TestSuite{
		serviceInstance:     serviceInstance,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		k8sChecker:          k8sResourcesChecker,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeoutSeconds),
	}
}

func (ts *TestSuite) CreateTestNamespace(t *testing.T) {
	namespace, err := ts.k8sClient.CreateNamespace()
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
			Namespace: testNamespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{Name: "redis"},
			ClusterServicePlanRef:  &v1beta1.ClusterObjectReference{Name: "micro"},
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
		return ts.helmClient.IsInstalled(ts.serviceInstance)
	})
	require.NoError(t, err, "Received timeout while waiting for release to install")
}

func (ts *TestSuite) WaitForReleaseToUninstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, ts.helmReleaseNotExist)
	require.NoError(t, err, "Received timeout while waiting for release to uninstall")
}

func (ts *TestSuite) CheckK8sResourcesDeployed(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.CheckK8sResources(t, ts.checkResourceDeployed)
}

func (ts *TestSuite) CheckK8sResourceRemoved(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.CheckK8sResources(t, ts.checkResourceRemoved)
}

func (ts *TestSuite) CleanUp() {
	//errors are not handled because resources can be deleted already
	ts.k8sClient.DeleteServiceInstance(ts.serviceInstance)
	ts.k8sClient.DeleteNamespace()
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	return !ts.helmClient.IsInstalled(ts.serviceInstance)
}

func (ts *TestSuite) checkResourceDeployed(resource interface{}, err error) bool {
	if err != nil {
		return false
	}

	return true
}

func (ts *TestSuite) checkResourceRemoved(_ interface{}, err error) bool {
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return true
		}
	}

	return false
}
