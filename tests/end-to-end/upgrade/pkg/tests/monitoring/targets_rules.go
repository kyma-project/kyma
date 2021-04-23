package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/monitoring/prom"
)

const prometheusURL = "http://monitoring-prometheus.kyma-system:9090"
const namespace = "kyma-system"
const expectedAlertManagers = 1
const expectedPrometheusInstances = 1
const expectedKubeStateMetrics = 1
const expectedGrafanaInstance = 1

// TargetsAndRulesTest checks that all targets and rules are healthy
type TargetsAndRulesTest struct {
	k8sCli        kubernetes.Interface
	monitoringCli *monitoringv1.MonitoringV1Client
	httpClient    *http.Client
}

// NewTargetsAndRulesTest creates a new instance of TargetsAndRulesTest
func NewTargetsAndRulesTest(k8sCli kubernetes.Interface, monitoringCli *monitoringv1.MonitoringV1Client) TargetsAndRulesTest {
	return TargetsAndRulesTest{
		k8sCli:        k8sCli,
		monitoringCli: monitoringCli,
		httpClient:    getHttpClient(),
	}
}

// CreateResources checks that all targets and rules are healthy and no alerts are firing before upgrade
func (t TargetsAndRulesTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("checking that monitoring pods are ready")
	if err := t.testPodsAreReady(); err != nil {
		return err
	}
	log.Println("checking that all targets are healthy before upgrade")
	if err := t.testTargetsAreHealthy(); err != nil {
		return err
	}
	log.Println("checking that all scrape pools have active targets before upgrade")
	if err := t.checkScrapePools(); err != nil {
		return err
	}
	log.Println("checking that all rules are healthy before upgrade")
	if err := t.testRulesAreHealthy(); err != nil {
		return err
	}
	log.Println("checking that no alerts are firing before upgrade")
	if err := t.checkAlerts(); err != nil {
		return err
	}
	return nil
}

// TestResources checks that all targets and rules are healthy and no alerts are firing after upgrade
func (t TargetsAndRulesTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("checking that monitoring pods are ready")
	if err := t.testPodsAreReady(); err != nil {
		return err
	}
	log.Println("checking that all targets are healthy after upgrade")
	if err := t.testTargetsAreHealthy(); err != nil {
		return err
	}
	log.Println("checking that all scrape pools have active targets after upgrade")
	if err := t.checkScrapePools(); err != nil {
		return err
	}
	log.Println("checking that all rules are healthy after upgrade")
	if err := t.testRulesAreHealthy(); err != nil {
		return err
	}
	log.Println("checking that no alerts are firing after upgrade")
	if err := t.checkAlerts(); err != nil {
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
			url := fmt.Sprintf("%s/api/v1/targets?state=active", prometheusURL)
			respBody, statusCode, err := t.doGet(url)
			if err != nil {
				return errors.Wrap(err, "cannot query targets")
			}
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				return errors.Wrapf(err, "error unmarshalling response.\nResponse body: %s", respBody)
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

func shouldIgnoreTarget(target prom.TargetLabels) bool {
	jobsToBeIgnored := []string{
		// Note: These targets will be tested here: https://github.com/kyma-project/kyma/issues/6457
	}

	podsToBeIgnored := []string{
		// Ignore the pods that are created during tests.
		"-testsuite-",
		"test",
		"nodejs12-",
		"nodejs14-",
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

func (t TargetsAndRulesTest) checkScrapePools() error {
	scrapePools := make(map[string]struct{})
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeout:
			tick.Stop()
			timeoutMessage := "Unable to scrape targets in the following scrape pool(s):\n"
			for scrapePool := range scrapePools {
				timeoutMessage += fmt.Sprintf("- %s\n", scrapePool)
			}
			return errors.Errorf(timeoutMessage)
		case <-tick.C:
			var err error
			scrapePools, err = t.buildScrapePoolSet()
			if err != nil {
				return errors.Wrap(err, "error while building the scrape pool set")
			}
			var resp prom.TargetsResponse
			url := fmt.Sprintf("%s/api/v1/targets?state=active", prometheusURL)
			respBody, statusCode, err := t.doGet(url)
			if err != nil {
				return errors.Wrap(err, "cannot query targets")
			}
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				return errors.Wrapf(err, "error unmarshalling response.\nResponse body: %s", respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				return errors.Errorf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
			}
			activeTargets := resp.Data.ActiveTargets
			for _, target := range activeTargets {
				delete(scrapePools, target.ScrapePool)
			}
			if len(scrapePools) == 0 {
				return nil
			}
		}
	}

}

func (t TargetsAndRulesTest) buildScrapePoolSet() (map[string]struct{}, error) {
	scrapePools := make(map[string]struct{})

	serviceMonitors, err := t.monitoringCli.ServiceMonitors("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing service monitors")
	}
	for _, serviceMonitor := range serviceMonitors.Items {
		if shouldIgnoreServiceMonitor(serviceMonitor.Name) {
			continue
		}
		for i := range serviceMonitor.Spec.Endpoints {
			scrapePool := fmt.Sprintf("%s/%s/%d", serviceMonitor.ObjectMeta.Namespace, serviceMonitor.Name, i)
			scrapePools[scrapePool] = struct{}{}
		}
	}

	podMonitors, err := t.monitoringCli.PodMonitors("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing pod monitors")
	}
	for _, podMonitor := range podMonitors.Items {
		if shouldIgnorePodMonitor(podMonitor.Name) {
			continue
		}
		for i := range podMonitor.Spec.PodMetricsEndpoints {
			scrapePool := fmt.Sprintf("%s/%s/%d", podMonitor.ObjectMeta.Namespace, podMonitor.Name, i)
			scrapePools[scrapePool] = struct{}{}
		}
	}

	return scrapePools, nil
}

func shouldIgnoreServiceMonitor(serviceMonitorName string) bool {
	var serviceMonitorsToBeIgnored = []string{
		// istio-mixer needs to be ignored until istio is upgraded to 1.5 in the latest release
		"istio-mixer",
		// monitoring-kube-proxy needs to be ignored until the fix for this issue https://github.com/kyma-project/kyma/issues/9457 is included in the latest release
		"monitoring-kube-proxy",
		// kiali-operator-metrics is created automatically by kiali operator and can't be disabled
		"kiali-operator-metrics",
		// tracing-metrics is created automatically by jaeger operator and can't be disabled
		"tracing-metrics",
	}

	for _, sm := range serviceMonitorsToBeIgnored {
		if sm == serviceMonitorName {
			return true
		}
	}
	return false
}

func shouldIgnorePodMonitor(podMonitorName string) bool {
	var podMonitorsToBeIgnored = []string{
		// The targets scraped by these podmonitors will be tested here: https://github.com/kyma-project/kyma/issues/6457
	}

	for _, pm := range podMonitorsToBeIgnored {
		if pm == podMonitorName {
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
			var resp prom.RulesResponse
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
			rulesGroups := resp.Data.Groups
			timeoutMessage = ""
			for _, group := range rulesGroups {
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

func (t TargetsAndRulesTest) checkAlerts() error {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			return errors.Errorf(timeoutMessage)
		case <-tick.C:
			var resp prom.AlertsResponse
			url := fmt.Sprintf("%s/api/v1/alerts", prometheusURL)
			respBody, statusCode, err := t.doGet(url)
			if err != nil {
				return errors.Wrap(err, "cannot query alerts")
			}
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				return errors.Wrapf(err, "error unmarshalling response. Response body: %s", respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				return errors.Errorf("error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
			}
			noFiringAlerts := true
			alerts := resp.Data.Alerts
			timeoutMessage = ""
			for _, alert := range alerts {
				if shouldIgnoreAlert(alert) {
					continue
				}
				if alert.State == "firing" {
					noFiringAlerts = false
					timeoutMessage += fmt.Sprintf("Alert with name=%s is firing\n", alert.Labels.AlertName)
				}
			}
			if noFiringAlerts {
				return nil
			}
		}
	}
}

func shouldIgnoreAlert(alert prom.Alert) bool {
	if alert.Labels.Severity != "critical" {
		return true
	}

	var alertNamesToIgnore = []string{
		// Watchdog is an alert meant to ensure that the entire alerting pipeline is functional, so it should always be firing,
		"Watchdog",
		// Scrape limits can be exceeded on long-running clusters and can be ignored
		"ScrapeLimitForTargetExceeded",
	}

	for _, name := range alertNamesToIgnore {
		if name == alert.Labels.AlertName {
			return true
		}
	}

	return false
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
