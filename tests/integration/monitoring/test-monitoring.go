package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type DataUp struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

type Result struct {
	Metric Metric `json:"metric"`
}
type ResponseUp struct {
	DataUp    DataUp `json:"data"`
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

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

type Metric struct {
	Name      string `json:"__name__"`
	Endpoint  string `json:"endpoint"`
	Instance  string `json:"instance"`
	Job       string `json:"job"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Service   string `json:"service"`
}

const prometheusURL string = "http://monitoring-prometheus.kyma-system:9090"
const grafanaURL string = "http://monitoring-grafana.kyma-system"
const namespace = "kyma-system"
const expectedAlertManagers = 1
const expectedPrometheusInstances = 1

const expectedKubeStateMetrics = 1
const expectedGrafanaInstance = 1

func getNumberofNodeExporter() int {
	cmd := exec.Command("kubectl", "get", "nodes")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while kubectl get nodes: %v.\nKubectl output: %s.", err, stdoutStderr)
	}
	outputArr := strings.Split(string(stdoutStderr), "\n")
	return len(outputArr) - 2
}

func main() {
	testPodsAreReady()
	testQueryTargets(prometheusURL)
	testGrafanaIsReady(grafanaURL)
	checkLambdaUIDashboard()

	log.Printf("Monitoring tests are successful!")
}

func testGrafanaIsReady(url string) {
	if b, statusCode := doGet(url); statusCode != 302 {
		log.Fatalf("Test grafana: Expected HTTP response code 302 but got %v.\nResponse body: %s", statusCode, b)
	}
	log.Printf("Test grafana UI: Success")
}

func testQueryTargets(url string) {

	var respObj ResponseTargets
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	path := "/api/v1/targets"
	url = url + path
	expectedNodeExporter := getNumberofNodeExporter()
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
					case "alertmanager":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) && (respObj.DataTargets.ActiveTargets[index].Labels["pod"] == "alertmanager-monitoring-alertmanager-0") {
							actualAlertManagers += 1
						}
					case "prometheus":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) && (respObj.DataTargets.ActiveTargets[index].Labels["pod"] == "prometheus-monitoring-prometheus-0") {
							actualPrometheusInstances += 1
						}
					case "node-exporter":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) {
							actualNodeExporter += 1
						}
					case "kube-state":
						if isHealthy(respObj.DataTargets.ActiveTargets[index]) {
							actualKubeStateMetrics += 1
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

func isHealthy(activeTarget ActiveTarget) bool {
	if activeTarget.Health == "up" {
		return true
	}
	return false
}

func testPodsAreReady() {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	expectedNodeExporter := getNumberofNodeExporter()
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
			cmd := exec.Command("kubectl", "get", "pods", "-l", "app in (alertmanager,prometheus,grafana,prometheus-node-exporter)", "-n", namespace, "--no-headers")
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error while kubectl get pods: %s.Kubectl output: %s. ", err, stdoutStderr)
			}
			outputArr := strings.Split(string(stdoutStderr), "\n")
			for index := range outputArr {
				if len(outputArr[index]) != 0 {
					podName, isReady := getPodStatus(string(outputArr[index]))
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
			}

			cmd = exec.Command("kubectl", "get", "pods", "-l", "app.kubernetes.io/name=kube-state-metrics", "-n", namespace, "--no-headers")
			stdoutStderr, err = cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error while kubectl get pods: %s.Kubectl output: %s. ", err, stdoutStderr)
			}
			outputArr = strings.Split(string(stdoutStderr), "\n")
			for index := range outputArr {
				if len(outputArr[index]) != 0 {
					podName, isReady := getPodStatus(string(outputArr[index]))
					if isReady {
						switch true {
						case strings.Contains(podName, "kube-state-metrics"):
							actualKubeStateMetrics++
						}
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

func getPodStatus(stdout string) (podName string, isReady bool) {
	isReady = false
	stdoutArr := regexp.MustCompile("( )+").Split(stdout, -1)
	podName = stdoutArr[0]
	readyCount := strings.Split(stdoutArr[1], "/")
	status := stdoutArr[2]
	if strings.ToUpper(status) == "RUNNING" && readyCount[0] == readyCount[1] {
		isReady = true
	}
	return
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
