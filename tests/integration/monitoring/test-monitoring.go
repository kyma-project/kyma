package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	"k8s.io/client-go/kubernetes"
)

type ResponseTargets struct {
	DataTargets DataTargets `json:"data"`
	Status      string      `json:"status"`
	ErrorType   string      `json:"errorType"`
	Error       string      `json:"error"`
}

type DataTargets struct {
	ActiveTargets []ActiveTarget `json:"activeTargets"`
}

type ActiveTarget struct {
	DiscoveredLabels map[string]string `json:"discoveredLabels"`
	Labels           map[string]string `json:"labels"`
	ScrapeURL        string            `json:"scrapeUrl"`
	Health           string            `json:"health"`
}

type monitoringTest struct {
	coreClient *kubernetes.Clientset
}

const prometheusURL string = "http://monitoring-prometheus.kyma-system:9090"
const grafanaURL string = "http://monitoring-grafana.kyma-system"
const namespace = "kyma-system"
const expectedAlertManagers = 1
const expectedPrometheusInstances = 1
const expectedKubeStateMetrics = 1
const expectedGrafanaInstance = 1

func main() {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		log.Fatalf("Cannot create REST config, err: %v", err)
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Cannot create client config, err: %v", err)
	}

	mt := &monitoringTest{coreClient: coreClient}

	mt.testPodsAreReady()
	mt.testQueryTargets(prometheusURL)
	mt.testGrafanaIsReady(grafanaURL)
	mt.checkLambdaUIDashboard()

	log.Printf("Monitoring tests are successful!")
}

func (mt *monitoringTest) testPodsAreReady() {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	expectedNodeExporter := mt.getNumberofNodeExporter()
	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		actualGrafanaInstance := 0
		select {
		case <-timeout:
			log.Println("Timed out: pods are still unready!!")
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

		case <-tick:
			pods, err := mt.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app in (alertmanager,prometheus,prometheus-node-exporter)"})
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

			pods, err = mt.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=kube-state-metrics"})
			if err != nil {
				log.Fatalf("Error while kubectl get pods, err: %v", err)
			}

			for _, pod := range pods.Items {
				podName := pod.Name
				isReady := getPodStatus(pod)
				if isReady {
					switch true {
					case strings.Contains(podName, "kube-state-metrics"):
						actualKubeStateMetrics++
					}
				}
			}

			if expectedAlertManagers == actualAlertManagers && expectedNodeExporter == actualNodeExporter && expectedPrometheusInstances == actualPrometheusInstances && expectedKubeStateMetrics == actualKubeStateMetrics && expectedGrafanaInstance == actualGrafanaInstance {
				log.Println("Test pods status: All pods are ready!!")
				return
			}
			log.Println("Waiting for the pods to be READY!!")
		}
	}
}

func (mt *monitoringTest) testQueryTargets(url string) {

	var respObj ResponseTargets
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	path := "/api/v1/targets"
	url = url + path
	expectedNodeExporter := mt.getNumberofNodeExporter()
	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		select {
		case <-timeout:
			log.Printf("Test prometheus API %v: result: Timed out!!", url)
			if expectedAlertManagers != actualAlertManagers {
				log.Fatalf("Expected alertmanager healthy is %d but got %d instances", expectedAlertManagers, actualAlertManagers)
			}
			if expectedNodeExporter != actualNodeExporter {
				log.Fatalf("Expected node exporter healthy is %d but got %d instances", expectedNodeExporter, actualNodeExporter)
			}
			if expectedPrometheusInstances != actualPrometheusInstances {
				log.Fatalf("Expected prometheus healthy is %d but got %d instances", expectedPrometheusInstances, actualPrometheusInstances)
			}
			if expectedKubeStateMetrics != actualKubeStateMetrics {
				log.Fatalf("Expected kube-state-metrics healthy is %d but got %d instances", expectedKubeStateMetrics, actualKubeStateMetrics)
			}
		case <-tick:
			respBody, statusCode := doGet(url)
			err := json.Unmarshal([]byte(respBody), &respObj)
			if err != nil {
				log.Fatalf("Error unmarshalling response: %v.\n Response body: %s", err, respBody)
			}
			if statusCode == 200 && respObj.Status != "success" {
				log.Fatalf("Error in response status with errorType: %s error: %s", respObj.Error, respObj.ErrorType)
			}

			for index := range respObj.DataTargets.ActiveTargets {
				if val, ok := respObj.DataTargets.ActiveTargets[index].Labels["job"]; ok {
					switch val {
					case "monitoring-alertmanager":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) && (respObj.DataTargets.ActiveTargets[index].Labels["pod"] == "alertmanager-monitoring-alertmanager-0") {
							actualAlertManagers++
						}
					case "monitoring-prometheus":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) && (respObj.DataTargets.ActiveTargets[index].Labels["pod"] == "prometheus-monitoring-prometheus-0") {
							actualPrometheusInstances++
						}
					case "node-exporter":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) {
							actualNodeExporter++
						}
					case "kube-state-metrics":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) {
							actualKubeStateMetrics++
						}
					}
				}
			}
			if expectedAlertManagers == actualAlertManagers && expectedNodeExporter == actualNodeExporter && expectedPrometheusInstances == actualPrometheusInstances && expectedKubeStateMetrics == actualKubeStateMetrics {
				log.Printf("Test prometheus API %v: result: All pods are healthy!!", url)
				return
			}
			log.Printf("Waiting for all instances to be healthy!!")
		}
	}
}

func (mt *monitoringTest) testGrafanaIsReady(url string) {
	if b, statusCode := doGet(url); statusCode != 302 {
		log.Fatalf("Test grafana: Expected HTTP response code 302 but got %v.\nResponse body: %s", statusCode, b)
	}
	log.Printf("Test grafana UI: Success")
}

func (mt *monitoringTest) getNumberofNodeExporter() int {
	nodes, err := mt.coreClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error while listing the nodes, err: %v", err)
	}

	return len(nodes.Items)
}

func isHealthy(activeTarget ActiveTarget) bool {
	return activeTarget.Health == "up"
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

func doGet(url string) (string, int) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("HTTP GET call fails: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	code := resp.StatusCode
	return string(body), code
}
