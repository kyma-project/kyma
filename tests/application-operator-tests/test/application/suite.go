package application

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/release"

	"k8s.io/apimachinery/pkg/labels"

	v1 "k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	testAppName          = "operator-app-test-%s"
	defaultCheckInterval = 2 * time.Second

	releaseLabelKey  = "release"
	helmTestLabelKey = "helm-chart-test"
)

type TestSuite struct {
	application string
	ctx         context.Context

	config     testkit.TestConfig
	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient

	installationTimeout time.Duration
	labelSelector       string
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	app := fmt.Sprintf(testAppName, rand.String(4))

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient, err := testkit.NewHelmClient(config.HelmDriver)
	require.NoError(t, err)

	testPodsLabels := labels.Set{
		releaseLabelKey:  app,
		helmTestLabelKey: "true",
	}

	return &TestSuite{
		application:         app,
		ctx:                 context.Background(),
		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeoutSeconds),
		labelSelector:       labels.SelectorFromSet(testPodsLabels).String(),
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	t.Logf("Creating %s Application", ts.application)
	_, err := ts.k8sClient.CreateDummyApplication(ts.ctx, ts.application, ts.application, false)
	require.NoError(t, err)
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	t.Log("Cleaning up...")
	err := ts.k8sClient.DeleteApplication(ts.ctx, ts.application, metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) WaitForApplicationToBeDeployed(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		app, err := ts.k8sClient.GetApplication(ts.ctx, ts.application, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return app.Status.InstallationStatus.Status == release.StatusDeployed.String()
	})

	require.NoError(t, err)
}

func (ts *TestSuite) RunApplicationTests(t *testing.T) {

	t.Log("Running application tests")
	_, err := ts.helmClient.TestRelease(ts.application, ts.config.Namespace)

	t.Logf("%s tests finished.", ts.application)

	ts.getLogsAndCleanup(t)

	if err != nil {
		t.Errorf("Error while executing tests for %s release: %s", ts.application, err.Error())
		t.Logf("%s tests failed", ts.application)
		t.Fatal("Application tests failed")
	}
}

func (ts *TestSuite) getLogsAndCleanup(t *testing.T) {
	podsToFetch, err := ts.k8sClient.ListPods(ts.ctx, metav1.ListOptions{LabelSelector: ts.labelSelector})
	require.NoError(t, err)

	for _, pod := range podsToFetch.Items {
		ts.getPodLogs(t, pod)
		ts.deleteTestPod(t, pod)
	}
}

func (ts *TestSuite) getPodLogs(t *testing.T, pod v1.Pod) {

	var testContainer string
	for _, c := range pod.Spec.Containers {
		if strings.HasPrefix(c.Name, ts.application) {
			testContainer = c.Name
		}
	}

	req := ts.k8sClient.GetLogs(ts.ctx, pod.Name, &v1.PodLogOptions{
		Container: testContainer,
	})

	reader, err := req.Stream(ts.ctx)
	require.NoError(t, err)

	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	strLogs := strings.Replace(string(bytes), "\t", "    ", -1)

	t.Logf("--------------------------------------------Logs from %s test--------------------------------------------", pod.Name)
	lines := strings.Split(strLogs, "\n")
	for _, l := range lines {
		t.Log(l)
	}
	t.Logf("--------------------------------------------End of logs from %s test--------------------------------------------", pod.Name)
}

func (ts *TestSuite) deleteTestPod(t *testing.T, pod v1.Pod) {
	err := ts.k8sClient.DeletePod(ts.ctx, pod.Name, metav1.DeleteOptions{})
	require.NoError(t, err)
}
