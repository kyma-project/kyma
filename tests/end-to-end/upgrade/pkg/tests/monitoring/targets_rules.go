package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/monitoring/prom"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const prometheusURL = "http://monitoring-prometheus.kyma-system:9090"
const namespace = "kyma-system"
const expectedAlertManagers = 1
const expectedPrometheusInstances = 1
const expectedKubeStateMetrics = 1
const expectedGrafanaInstance = 1

// TargetsAndRulesTest checks that all targets and rules are healthy
type TargetsAndRulesTest struct {
	k8sCli     kubernetes.Interface
	httpClient *http.Client
}

// NewTargetsAndRulesTest creates a new instance of TargetsAndRulesTest
func NewTargetsAndRulesTest(k8sCli kubernetes.Interface) TargetsAndRulesTest {
	return TargetsAndRulesTest{
		k8sCli:     k8sCli,
		httpClient: getHttpClient(),
	}
}

// CreateResources checks that all targets and rules are healthy before upgrade
func (t TargetsAndRulesTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("checking that monitoring pods are ready")
	if err := t.testPodsAreReady(); err != nil {
		return err
	}
	log.Println("checking that all rules are healthy before upgrade")
	if err := t.testRulesAreHealthy(); err != nil {
		return err
	}
	return nil
}

// CreateResources checks that all targets and rules are healthy after upgrade
func (t TargetsAndRulesTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("checking that monitoring pods are ready")
	if err := t.testPodsAreReady(); err != nil {
		return err
	}
	log.Println("checking that all rules are healthy after upgrade")
	if err := t.testRulesAreHealthy(); err != nil {
		return err
	}
	return nil
}

func (t TargetsAndRulesTest) testPodsAreReady() error {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	expectedNodeExporter, err := t.getNumberofNodeExporter()
	if err != nil {
		return errors.Wrap(err, "cannot get number of nodes")
	}
	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		actualGrafanaInstance := 0
		select {
		case <-timeout:
			tick.Stop()
			if expectedAlertManagers != actualAlertManagers {
				return errors.Errorf("timed out: expected alertmanager running is %d but got %d instances", expectedAlertManagers, actualAlertManagers)
			}
			if expectedNodeExporter != actualNodeExporter {
				return errors.Errorf("timed out: expected node exporter running is %d but got %d instances", expectedNodeExporter, actualNodeExporter)
			}
			if expectedPrometheusInstances != actualPrometheusInstances {
				return errors.Errorf("timed out: expected prometheus running is %d but got %d instances", expectedPrometheusInstances, actualPrometheusInstances)
			}
			if expectedKubeStateMetrics != actualKubeStateMetrics {
				return errors.Errorf("timed out: expected kube-state-metrics running is %d but got %d instances", expectedKubeStateMetrics, actualKubeStateMetrics)
			}
			if expectedGrafanaInstance != actualGrafanaInstance {
				return errors.Errorf("timed out: expected grafana running is %d but got %d instances", expectedGrafanaInstance, actualGrafanaInstance)
			}

		case <-tick.C:
			pods, err := t.k8sCli.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app in (alertmanager,prometheus,grafana,prometheus-node-exporter)"})
			if err != nil {
				return errors.Wrapf(err, "error while kubectl get pods")
			}

			for _, pod := range pods.Items {
				podName := pod.Name
				isReady := getPodStatus(pod)
				if isReady {
					switch true {
					case strings.Contains(podName, "alertmanager"):
						actualAlertManagers++

					case strings.Contains(podName, "node-exporter"):
						actualNodeExporter++

					case strings.Contains(podName, "prometheus-monitoring"):
						actualPrometheusInstances++

					case strings.Contains(podName, "grafana"):
						actualGrafanaInstance++
					}
				}
			}

			pods, err = t.k8sCli.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=kube-state-metrics"})
			if err != nil {
				return errors.Wrapf(err, "error while kubectl get pods")
			}

			for _, pod := range pods.Items {
				podName := pod.Name
				isReady := getPodStatus(pod)
				if isReady && strings.Contains(podName, "kube-state-metrics") {
					actualKubeStateMetrics++
				}
			}

			if expectedAlertManagers == actualAlertManagers && expectedNodeExporter == actualNodeExporter && expectedPrometheusInstances == actualPrometheusInstances && expectedKubeStateMetrics == actualKubeStateMetrics && expectedGrafanaInstance == actualGrafanaInstance {
				return nil
			}
		}
	}
}

func (t TargetsAndRulesTest) testTargetsAreHealthy() error {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf(timeoutMessage)
		case <-tick.C:
			var resp prom.TargetsResponse
			url := fmt.Sprintf("%s/api/v1/targets", prometheusURL)
			respBody, statusCode, err := t.doGet(url)
			if err != nil {
				return errors.Wrap(err, "cannot query targets")
			}
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				return errors.Wrapf(err, "error unmarshalling response. Response body: %s", respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				return errors.Errorf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
			}
			activeTargets := resp.Data.ActiveTargets
			allTargetsAreHealthy := true
			timeoutMessage = ""
			for _, target := range activeTargets {
				if shouldIgnoreTarget(target.Labels) {
					continue
				}
				if target.Health != "up" {
					allTargetsAreHealthy = false
					timeoutMessage += "The following target is not healthy:\n"
					for label, value := range target.Labels {
						timeoutMessage += fmt.Sprintf("- %s=%s\n", label, value)
					}
					timeoutMessage += fmt.Sprintf("- errorMessage: %s", target.LastError)
				}
			}
			if allTargetsAreHealthy {
				return nil
			}
		}
	}

}

func shouldIgnoreTarget(target prom.Labels) bool {
	jobsToBeIgnored := []string{
		// Note: These targets will be tested here: https://github.com/kyma-project/kyma/issues/6457
		"knative-eventing/knative-eventing-event-mesh-dashboard-broker",
		"knative-eventing/knative-eventing-event-mesh-dashboard-httpsource",
	}

	podsToBeIgnored := []string{
		// Ignore the pods that are created during tests.
		"-testsuite-",
		"test",
		"nodejs12-",
		"nodejs10-",
		"upgrade",
		// Ignore the pods created by jobs which are executed after installation of control-plane.
		"compass-migration",
		"compass-director-tenant-loader-default",
		"compass-agent-configuration",
	}

	namespacesToBeIgnored := []string{"test", "e2e"} // Since some namespaces are named -e2e and some are e2e-

	for _, p := range podsToBeIgnored {
		if strings.Contains(target["pod"], p) {
			return true
		}
	}

	for _, j := range jobsToBeIgnored {
		if target["job"] == j {
			return true
		}
	}

	for _, n := range namespacesToBeIgnored {
		if strings.Contains(target["namespace"], n) {
			return true
		}
	}

	return false
}

func (t TargetsAndRulesTest) testRulesAreHealthy() error {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf(timeoutMessage)
		case <-tick.C:
			var resp prom.AlertResponse
			url := fmt.Sprintf("%s/api/v1/rules", prometheusURL)
			respBody, statusCode, err := t.doGet(url)
			if err != nil {
				return errors.Wrap(err, "cannot query rules")
			}
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				return errors.Wrapf(err, "error unmarshalling response. Response body: %s", respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				return errors.Errorf("error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
			}
			allRulesAreHealthy := true
			alertDataGroups := resp.Data.Groups
			timeoutMessage = ""
			for _, group := range alertDataGroups {
				for _, rule := range group.Rules {
					if rule.Health != "ok" {
						allRulesAreHealthy = false
						timeoutMessage += fmt.Sprintf("Rule with name=%s is not healthy\n", rule.Name)
					}
				}
			}
			if allRulesAreHealthy {
				return nil
			}
		}
	}

}

func (t TargetsAndRulesTest) getNumberofNodeExporter() (int, error) {
	nodes, err := t.k8sCli.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return -1, errors.Wrap(err, "error while listing the nodes")
	}

	return len(nodes.Items), nil
}

func getPodStatus(pod corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return true
}

func getHttpClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	return client
}

func (t TargetsAndRulesTest) doGet(url string) (string, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", -1, errors.Wrap(err, "cannot create a new HTTP request")
	}
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", -1, errors.Wrapf(err, "cannot send HTTP request to %s", url)
	}
	defer resp.Body.Close()
	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		return "", -1, errors.Wrap(err, "cannot read response body")
	}
	return body.String(), resp.StatusCode, nil
}
