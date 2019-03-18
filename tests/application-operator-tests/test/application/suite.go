package application

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	hapirelease1 "k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

const (
	testAppName          = "ctrl-app-test-%s"
	defaultCheckInterval = 2 * time.Second
)

type TestSuite struct {
	application string

	config     testkit.TestConfig
	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient

	installationTimeout time.Duration
	labelSelector       string
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	app := fmt.Sprintf(testAppName, rand.String(5))

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost)

	testPodsLabels := labels.Set{
		"release":         app,
		"helm-chart-test": "true",
	}

	return &TestSuite{
		application: app,

		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeout),
		labelSelector:       labels.SelectorFromSet(testPodsLabels).String(),
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	log.Infof("Creating %s Application", ts.application)
	_, err := ts.k8sClient.CreateDummyApplication(ts.application, ts.application, false)
	require.NoError(t, err)
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	log.Info("Cleaning up...")
	err := ts.k8sClient.DeleteApplication(ts.application, &metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) RunApplicationTests(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	log.Info("Running application tests")
	responseChan, errorChan := ts.helmClient.TestRelease(ts.application)

	go func(responseChan <-chan *rls.TestReleaseResponse) {
		defer wg.Done()

		testFailed := false

		for msg := range responseChan {
			log.Infoln(msg.String())
			if msg.Status == hapirelease1.TestRun_FAILURE {
				testFailed = true
			}
		}

		if testFailed {
			log.Infof("%s tests failed", ts.application)
			ts.GetTestPodsLogs(t)
		}

		log.Infof("%s tests finished. Message channel closed", ts.application)
	}(responseChan)

	go func(errorChan <-chan error) {
		defer wg.Done()

		for err := range errorChan {
			log.Errorf(err.Error())
			t.Fatalf("Error while executing tests for %s release", ts.application)
		}

		log.Infoln("Error channel closed")
	}(errorChan)

	wg.Wait()

}

func (ts *TestSuite) WaitForApplicationToBeDeployed(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		app, err := ts.k8sClient.GetApplication(ts.application, metav1.GetOptions{})
		if err != nil {
			return false
		}

		if app.Status.InstallationStatus.Status != hapirelease1.Status_DEPLOYED.String() {
			return false
		}

		return true
	})

	require.NoError(t, err)
}

func (ts *TestSuite) GetTestPodsLogs(t *testing.T) {
	podsToFetch, err := ts.k8sClient.ListPods(metav1.ListOptions{LabelSelector: ts.labelSelector})
	require.NoError(t, err)

	log.Infoln(podsToFetch)

	for _, pod := range podsToFetch.Items {
		log.Infoln(pod.Name)
		ts.getPodLogs(t, pod)
	}
}

func (ts *TestSuite) getPodLogs(t *testing.T, pod v1.Pod) {
	req := ts.k8sClient.GetLogs(pod.Name, &v1.PodLogOptions{})

	reader, err := req.Stream()
	require.NoError(t, err)

	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	strLogs := strings.Replace(string(bytes), "\t", "    ", -1)

	log.Infof("--------------------------------------------Logs from %s test--------------------------------------------", pod.Name)
	lines := strings.Split(strLogs, "\n")
	for _, l := range lines {
		log.Infoln(l)
	}
	log.Infof("--------------------------------------------End of logs from %s test--------------------------------------------", pod.Name)
}
