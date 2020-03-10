package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
)

type ResponseTargets struct {
	DataTargets []DataTarget `json:"data"`
	Status      string       `json:"status"`
	ErrorType   string       `json:"errorType"`
	Error       string       `json:"error"`
}

type DataTarget struct {
	Target TargetData `json:"target"`
}

type TargetData struct {
	Endpoint  string `json:endpoint`
	Instance  string `json:instancce`
	Job       string `json:"job"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Service   string `json:"service"`
}

type TestTargets struct {
	Targets []TestTarget
}

type TestTarget struct {
	Job    string
	Metric string
	Count  int
}

const prometheusURL string = "http://localhost:9090"
const grafanaURL string = "http://monitoring-grafana.kyma-system"
const namespace = "kyma-system"
const expectedAlertManagers = 1
const expectedPrometheusInstances = 1
const expectedKubeStateMetrics = 1
const expectedGrafanaInstance = 1

var kubeConfig *rest.Config
var k8sClient *kubernetes.Clientset

func main() {
	kubeConfig = loadKubeConfigOrDie()
	k8sClient = kubernetes.NewForConfigOrDie(kubeConfig)

	testPodsAreReady()
	testQueryTargets()
	testGrafanaIsReady(grafanaURL)
	checkLambdaUIDashboard()

	log.Printf("Monitoring tests are successful!")
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

		case <-tick:
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
				if isReady {
					switch true {
					case strings.Contains(podName, "kube-state-metrics"):
						actualKubeStateMetrics++
					}
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

func testQueryTargets() {

	var respObj ResponseTargets
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	path := `%s/api/v1/targets/metadata?match_target={job="%s"}&metric=%s&limit=%d`
	expectedNodeExporter := getNumberofNodeExporter()
	var testTargets = []TestTarget{
		{"monitoring-alertmanager", "go_goroutines", expectedAlertManagers},
		{"monitoring-prometheus", "go_goroutines", expectedPrometheusInstances},
		{"node-exporter", "go_goroutines", expectedNodeExporter},
		{"kube-state-metrics", "kube_daemonset_status_current_number_scheduled", expectedKubeStateMetrics}}

	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		select {
		case <-timeout:
			log.Printf("Test prometheus API %v: result: Timed out!!", prometheusURL)
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

			for _, val := range testTargets {
				var url = fmt.Sprintf(path, prometheusURL, val.Job, val.Metric, val.Count)
				respBody, statusCode := doGet(url)
				err := json.Unmarshal([]byte(respBody), &respObj)
				if err != nil {
					log.Fatalf("Error unmarshalling response: %v.\n Response body: %s", err, respBody)
				}
				if statusCode != 200 || respObj.Status != "success" {
					log.Fatalf("Error in response status with errorType: %s error: %s for %s", respObj.Error, respObj.ErrorType, val.Job)
				}
				switch val.Job {
				case "monitoring-alertmanager":
					if respObj.DataTargets[0].Target.Pod == "alertmanager-monitoring-alertmanager-0" {
						actualAlertManagers++
					}
				case "monitoring-prometheus":
					if respObj.DataTargets[0].Target.Pod == "prometheus-monitoring-prometheus-0" {
						actualPrometheusInstances++
					}
				case "node-exporter":
					actualNodeExporter = len(respObj.DataTargets)
				case "kube-state-metrics":
					actualKubeStateMetrics++
				}

			}
			if expectedAlertManagers == actualAlertManagers && expectedNodeExporter == actualNodeExporter && expectedPrometheusInstances == actualPrometheusInstances && expectedKubeStateMetrics == actualKubeStateMetrics {
				log.Printf("Test prometheus API %v: result: All pods are healthy!!", prometheusURL)
				return
			}
			log.Printf("Waiting for all instances to be healthy!!")
		}
	}
}

func testGrafanaIsReady(url string) {
	if b, statusCode := doGet(url); statusCode != 302 {
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
