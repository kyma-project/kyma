package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/integration/monitoring/prom"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const prometheusURL = "http://monitoring-prometheus.kyma-system:9090"
const grafanaURL = "http://monitoring-grafana.kyma-system"
const namespace = "kyma-system"
const expectedAlertManagers = 1
const expectedPrometheusInstances = 1
const expectedKubeStateMetrics = 1
const expectedGrafanaInstance = 1

var kubeConfig *rest.Config
var k8sClient *kubernetes.Clientset
var monitoringClient *monitoringv1.MonitoringV1Client
var httpClient *http.Client

func main() {
	kubeConfig = loadKubeConfigOrDie()
	k8sClient = kubernetes.NewForConfigOrDie(kubeConfig)
	monitoringClient = monitoringv1.NewForConfigOrDie(kubeConfig)
	httpClient = getHttpClient()

	testPodsAreReady()
	testTargetsAreHealthy()
	checkScrapePools()
	testRulesAreHealthy()
	checkAlerts()
	testGrafanaIsReady()
	checkLambdaUIDashboard()

	log.Println("Monitoring tests are successful!")
}

func testPodsAreReady() {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	expectedNodeExporter := getNumberofNodeExporter()
	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		actualGrafanaInstance := 0
		select {
		case <-timeout:
			log.Println("Timed out: pods are still not ready")

			if expectedAlertManagers != actualAlertManagers {
				log.Fatalf("Expected alertmanager running is %d but got %d instances", expectedAlertManagers, actualAlertManagers)
			}
			if expectedNodeExporter != actualNodeExporter {
				log.Fatalf("Expected node exporter running is %d but got %d instances", expectedNodeExporter, actualNodeExporter)
			}
			if expectedPrometheusInstances != actualPrometheusInstances {
				log.Fatalf("Expected prometheus running is %d but got %d instances", expectedPrometheusInstances, actualPrometheusInstances)
			}
			if expectedKubeStateMetrics != actualKubeStateMetrics {
				log.Fatalf("Expected kube-state-metrics running is %d but got %d instances", expectedKubeStateMetrics, actualKubeStateMetrics)
			}
			if expectedGrafanaInstance != actualGrafanaInstance {
				log.Fatalf("Expected grafana running is %d but got %d instances", expectedGrafanaInstance, actualGrafanaInstance)
			}

		case <-tick.C:
			pods, err := k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app in (alertmanager,prometheus,grafana,prometheus-node-exporter)"})
			if err != nil {
				log.Fatalf("Error while kubectl get pods, err: %v", err)
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

			pods, err = k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=kube-state-metrics"})
			if err != nil {
				log.Fatalf("Error while kubectl get pods, err: %v", err)
			}

			for _, pod := range pods.Items {
				podName := pod.Name
				isReady := getPodStatus(pod)
				if isReady && strings.Contains(podName, "kube-state-metrics") {
					actualKubeStateMetrics++
				}
			}

			if expectedAlertManagers == actualAlertManagers && expectedNodeExporter == actualNodeExporter && expectedPrometheusInstances == actualPrometheusInstances && expectedKubeStateMetrics == actualKubeStateMetrics && expectedGrafanaInstance == actualGrafanaInstance {
				log.Println("Test pods status: All pods are ready!")
				return
			}
			log.Println("Waiting for the pods to be ready..")
		}
	}
}

func testTargetsAreHealthy() {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)

	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			log.Fatal(timeoutMessage)
		case <-tick.C:
			var resp prom.TargetsResponse
			url := fmt.Sprintf("%s/api/v1/targets?state=active", prometheusURL)
			respBody, statusCode := doGet(url)
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				log.Fatalf("Error unmarshalling response: %v.\nResponse body: %s", err, respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				log.Fatalf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
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
				log.Println("All targets are healthy")
				return
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

func checkScrapePools() {
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
			log.Fatal(timeoutMessage)
		case <-tick.C:
			scrapePools = buildScrapePoolSet()
			var resp prom.TargetsResponse
			url := fmt.Sprintf("%s/api/v1/targets?state=active", prometheusURL)
			respBody, statusCode := doGet(url)
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				log.Fatalf("Error unmarshalling response: %v.\nResponse body: %s", err, respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				log.Fatalf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
			}
			activeTargets := resp.Data.ActiveTargets
			for _, target := range activeTargets {
				delete(scrapePools, target.ScrapePool)
			}
			if len(scrapePools) == 0 {
				log.Println("All scrape pools have active targets")
				return
			}
		}
	}

}

func buildScrapePoolSet() map[string]struct{} {
	scrapePools := make(map[string]struct{})

	serviceMonitors, err := monitoringClient.ServiceMonitors("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error while listing service monitors: %v", err)
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

	podMonitors, err := monitoringClient.PodMonitors("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error while listing pod monitors: %v", err)
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

	return scrapePools
}

func shouldIgnoreServiceMonitor(serviceMonitorName string) bool {
	var serviceMonitorsToBeIgnored = []string{
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

func testRulesAreHealthy() {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			log.Fatal(timeoutMessage)
		case <-tick.C:
			var resp prom.RulesResponse
			url := fmt.Sprintf("%s/api/v1/rules", prometheusURL)
			respBody, statusCode := doGet(url)
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				log.Fatalf("Error unmarshalling response: %v.\nResponse body: %s", err, respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				log.Fatalf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
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
				log.Println("All rules are healthy")
				return
			}
		}
	}

}

func checkAlerts() {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			log.Fatal(timeoutMessage)
		case <-tick.C:
			var resp prom.AlertsResponse
			url := fmt.Sprintf("%s/api/v1/alerts", prometheusURL)
			respBody, statusCode := doGet(url)
			if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
				log.Fatalf("Error unmarshalling response: %v.\nResponse body: %s", err, respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				log.Fatalf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
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
				log.Println("No alerts are firing")
				return
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

func testGrafanaIsReady() {
	if b, statusCode := doGet(grafanaURL); statusCode != 302 {
		log.Fatalf("Test grafana: Expected HTTP response code 302 but got %v.\nResponse body: %s", statusCode, b)
	}
	log.Printf("Test grafana UI: Success")
}

func getNumberofNodeExporter() int {
	nodes, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error while listing the nodes, err: %v", err)
	}

	return len(nodes.Items)
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

func doGet(url string) (string, int) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Cannot create a new HTTP request: ", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("Cannot send HTTP request to %s: %v", url, err)
	}
	defer resp.Body.Close()
	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		log.Fatalf("Cannot read response body: %v", err)
	}
	return body.String(), resp.StatusCode
}

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Cannot create in-cluster config: %v", err)
		}
		return cfg
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Cannot read kubeconfig: %s", err)
	}
	return cfg
}
