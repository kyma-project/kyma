package monitoring

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	prometheusURL               = "http://monitoring-prometheus.kyma-system:9090"
	prometheusNamespace         = "kyma-system"
	expectedAlertManagers       = 1
	expectedPrometheusInstances = 1
	expectedKubeStateMetrics    = 1
)

type responseTargets struct {
	DataTargets dataTargets `json:"data"`
	Status      string      `json:"status"`
	ErrorType   string      `json:"errorType"`
	Error       string      `json:"error"`
}

type dataTargets struct {
	ActiveTargets []activeTarget `json:"activeTargets"`
}

type activeTarget struct {
	DiscoveredLabels map[string]string `json:"discoveredLabels"`
	Labels           map[string]string `json:"labels"`
	ScrapeURL        string            `json:"scrapeUrl"`
	Health           string            `json:"health"`
}

type prometheusTest struct {
	coreClient *kubernetes.Clientset
	log        logrus.FieldLogger
}

func NewPrometheusTest() (*prometheusTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return &prometheusTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return &prometheusTest{}, err
	}

	return &prometheusTest{
		coreClient: coreClient,
		log:        logrus.WithField("test", "prometheus"),
	}, nil
}

func (pt *prometheusTest) CreateResources(namespace string) {
	// There is no need to implement it for this test.
}

func (pt *prometheusTest) TestResources(namespace string) {
	pt.testPodsAreReady()
	pt.testQueryTargets(prometheusURL)
}

func (pt *prometheusTest) testPodsAreReady() {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	expectedNodeExporter := pt.getNumberofNodeExporter()
	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		select {
		case <-timeout:
			pt.log.Println("Timed out: pods are still not ready!")

			So(actualAlertManagers, ShouldEqual, expectedAlertManagers)
			So(actualNodeExporter, ShouldEqual, expectedNodeExporter)
			So(actualPrometheusInstances, ShouldEqual, expectedPrometheusInstances)
			So(actualKubeStateMetrics, ShouldEqual, expectedKubeStateMetrics)
		case <-tick:
			pods, err := pt.coreClient.CoreV1().Pods(prometheusNamespace).List(metav1.ListOptions{LabelSelector: "app in (alertmanager,prometheus,prometheus-node-exporter)"})
			So(err, ShouldBeNil)

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
					}
				}
			}

			pods, err = pt.coreClient.CoreV1().Pods(prometheusNamespace).List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=kube-state-metrics"})
			So(err, ShouldBeNil)

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

			if expectedAlertManagers == actualAlertManagers && expectedNodeExporter == actualNodeExporter && expectedPrometheusInstances == actualPrometheusInstances && expectedKubeStateMetrics == actualKubeStateMetrics {
				pt.log.Println("Test pods status: All pods are ready!")
				return
			}
			pt.log.Println("Waiting for the pods to be ready")
		}
	}
}

func (pt *prometheusTest) getNumberofNodeExporter() int {
	nodes, err := pt.coreClient.CoreV1().Nodes().List(metav1.ListOptions{})
	So(err, ShouldBeNil)

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

func (pt *prometheusTest) testQueryTargets(url string) {

	var respObj responseTargets
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	path := "/api/v1/targets"
	url = url + path
	expectedNodeExporter := pt.getNumberofNodeExporter()
	for {
		actualAlertManagers := 0
		actualPrometheusInstances := 0
		actualNodeExporter := 0
		actualKubeStateMetrics := 0
		select {
		case <-timeout:
			pt.log.Printf("Timed out: test prometheus API %v", url)
			So(actualAlertManagers, ShouldEqual, expectedAlertManagers)
			So(actualNodeExporter, ShouldEqual, expectedNodeExporter)
			So(actualPrometheusInstances, ShouldEqual, expectedPrometheusInstances)
			So(actualKubeStateMetrics, ShouldEqual, expectedKubeStateMetrics)
		case <-tick:
			respBody, statusCode := doGet(url)
			err := json.Unmarshal([]byte(respBody), &respObj)
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
			So(respObj.Status, ShouldEqual, "success")

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
				pt.log.Printf("Test prometheus API %v: result: All pods are healthy!", url)
				return
			}
			pt.log.Printf("Waiting for all instances to be healthy")
		}
	}
}

func isHealthy(activeTarget activeTarget) bool {
	return activeTarget.Health == "up"
}

func doGet(url string) (string, int) {
	req, err := http.NewRequest("GET", url, nil)
	So(err, ShouldBeNil)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	resp, err := client.Do(req)
	So(err, ShouldBeNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	So(err, ShouldBeNil)

	code := resp.StatusCode
	return string(body), code
}
