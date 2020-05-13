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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-project/kyma/tests/integration/monitoring/promAPI"
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
var httpClient *http.Client

func main() {
	kubeConfig = loadKubeConfigOrDie()
	k8sClient = kubernetes.NewForConfigOrDie(kubeConfig)
	httpClient = getHttpClient()

	testPodsAreReady()
	testTargetsAreHealthy()
	testRulesAreHealthy()
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
	tick := time.NewTicker(30 * time.Second)
	var labelsToBeIgnored = promAPI.Labels{
		"namespace": "e2e-event-mesh", // e2e test pod
	}
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			log.Fatal(timeoutMessage)
		case <-tick.C:
			var resp promAPI.TargetsResponse
			url := fmt.Sprintf("%s/api/v1/targets", prometheusURL)
			respBody, statusCode := doGet(url)
			err := json.Unmarshal([]byte(respBody), &resp)
			if err != nil {
				log.Fatalf("Error unmarshalling response: %v.\nResponse body: %s", err, respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				log.Fatalf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
			}
			activeTargets := resp.Data.ActiveTargets
			allTargetsAreHealthy := true
			timeoutMessage = ""
			for _, target := range activeTargets {
				// Ignoring the targets with certain labels
				if hasAnyLabel(target.Labels, labelsToBeIgnored) {
					continue
				}
				if target.Health != "up" {
					allTargetsAreHealthy = false
					timeoutMessage += fmt.Sprintf("Target with job=%s and instance=%s is not healthy\n", target.Labels["job"], target.Labels["instance"])
				}
			}
			if allTargetsAreHealthy {
				log.Println("All targets are healthy")
				return
			}
		}
	}

}

func hasAnyLabel(target, anyOf promAPI.Labels) bool {
	for l, _ := range target {
		if target[l] == anyOf[l] {
			return true
		}
	}
	return false
}

func testRulesAreHealthy() {
	timeout := time.After(3 * time.Minute)
	tick := time.NewTicker(30 * time.Second)
	var timeoutMessage string
	for {
		select {
		case <-timeout:
			tick.Stop()
			log.Fatal(timeoutMessage)
		case <-tick.C:
			var resp promAPI.AlertResponse
			url := fmt.Sprintf("%s/api/v1/rules", prometheusURL)
			respBody, statusCode := doGet(url)
			err := json.Unmarshal([]byte(respBody), &resp)
			if err != nil {
				log.Fatalf("Error unmarshalling response: %v.\nResponse body: %s", err, respBody)
			}
			if statusCode != 200 || resp.Status != "success" {
				log.Fatalf("Error in response status with ErrorType: %s.\nError: %s", resp.ErrorType, resp.Error)
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
				log.Println("All rules are healthy")
				return
			}
		}
	}

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
