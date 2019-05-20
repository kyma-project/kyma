package backupe2e

// Test http request to the prometheus api before backup and after restore.
// An expected result at a specific point in time is considered a succeeded test.
//
// {
//   "status":"success",
//   "data":{
//      "resultType":"vector",
//      "result":[
//         {
//           "metric":{},
//           "value":[
//               1551421406.195,
//               "1.661"
//            ]
//          }
//      ]
//    }
// }
//
// {success {vector [{map[] [1.551424874014e+09 1.661]}]}}

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	prometheusClient "github.com/coreos/prometheus-operator/pkg/client/versioned"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	domain                    = "http://monitoring-prometheus.kyma-system"
	prometheusNS              = "kyma-system"
	api                       = "/api/v1/query?"
	metricsQuery              = "max(sum(kube_pod_container_resource_requests_cpu_cores) by (instance))"
	port                      = "9090"
	metricName                = "kube_pod_container_resource_requests_cpu_cores"
	prometheusName            = "monitoring"
	prometheusPodName         = "prometheus-monitoring-0"
	prometheusServiceName     = "monitoring-prometheus"
	prometheusStatefulsetName = "prometheus-monitoring"
	prometheusPvcName         = "prometheus-monitoring-db-prometheus-monitoring-0"
	prometheusLabelSelector   = "app=prometheus"
)

type queryResponse struct {
	Status string       `json:"status"`
	Data   responseData `json:"data"`
}

type responseData struct {
	ResultType string       `json:"resultType"`
	Result     []dataResult `json:"result"`
}

type dataResult struct {
	Metric interface{}   `json:"metric,omitempty"`
	Value  []interface{} `json:"value,omitempty"`
}

type prometheusTest struct {
	metricName, uuid string
	coreClient       *kubernetes.Clientset
	prometheusClient *prometheusClient.Clientset
	response         queryResponse
	beforeBackup     bool
	expectedResult   string
	finalResult      string
	apiQuery
	pointInTime
	log logrus.FieldLogger
}

type pointInTime struct {
	floatValue      float64
	timeValue       time.Time
	formmattedValue string
}

type apiQuery struct {
	domain       string
	prometheusNS string
	api          string
	metricQuery  string
	port         string
}

var backedupPrometheus *v1.Prometheus

func NewPrometheusTest() (*prometheusTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return &prometheusTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return &prometheusTest{}, err
	}

	pClient, err := prometheusClient.NewForConfig(restConfig)
	if err != nil {
		return &prometheusTest{}, err
	}

	queryToApi := apiQuery{api: api, domain: domain, metricQuery: metricsQuery, port: port, prometheusNS: prometheusNS}

	return &prometheusTest{
		coreClient:       coreClient,
		prometheusClient: pClient,
		metricName:       metricName,
		uuid:             uuid.New().String(),
		apiQuery:         queryToApi,
		beforeBackup:     true,
		log:              logrus.WithField("test", "prometheus"),
	}, nil
}

func (point *pointInTime) pointInTime(f float64) {
	point.floatValue = f

	t := time.Unix(int64(f), 0) //gives unix time stamp in utc
	point.timeValue = t

	point.formmattedValue = t.Format(time.RFC3339)
}

type Connector interface {
	connectToPrometheusApi(domain, port, api, query, pointInTime string) error
}

func (qresp *queryResponse) connectToPrometheusApi(domain, port, api, query, pointInTime string) error {
	values := url.Values{}
	values.Set("query", query)
	if pointInTime != "" {
		values.Set("time", pointInTime)
	}

	uri := domain + ":" + port + api
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	url := uri + values.Encode()

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("http request to the api (%s) failed with '%s'", uri, err)
	}
	defer func() {
		err := resp.Body.Close()
		So(err, ShouldBeNil)
	}()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to get a reponse from the api. \n http response was '%s' (%d) and not OK (200). Body:\n%s\n", resp.Status, resp.StatusCode, string(body))
	}

	return qresp.decodeQueryResponse(body)

}

func whatIsThisThing(something interface{}) (float64, string, error) {
	switch i := something.(type) {
	case float64:
		return float64(i), "", nil
	case string:
		return float64(0), string(i), nil
	default:
		return float64(0), "", errors.New("unknown value is of incompatible type")
	}
}

func (qresp *queryResponse) decodeQueryResponse(jresponse []byte) error {
	err := json.Unmarshal(jresponse, &qresp)
	if err != nil {
		return fmt.Errorf("http response can't be Unmarshal: %v", err)
	}

	return nil
}

func (pt *prometheusTest) CreateResources(namespace string) {
	qresp := &queryResponse{}
	err := qresp.connectToPrometheusApi(pt.domain, pt.port, pt.api, pt.metricQuery, "")
	So(err, ShouldBeNil)

	pt.response = *qresp
	point := pointInTime{}
	if len(qresp.Data.Result) > 0 && len(qresp.Data.Result[0].Value) > 0 {
		values := qresp.Data.Result[0].Value
		for _, something := range values {

			f, s, err := whatIsThisThing(something)
			So(err, ShouldBeNil)

			if f != float64(0) {
				point.pointInTime(f)
				pt.pointInTime = point
			}

			if s != "" {
				pt.expectedResult = s
			}
		}

	}

}

func (pt *prometheusTest) TestResources(namespace string) {
	if !pt.beforeBackup {
		err := pt.deletePrometheus(prometheusNS, prometheusName)
		So(err, ShouldBeNil)

		err = pt.deletePod(prometheusNS, prometheusPodName, prometheusLabelSelector)
		So(err, ShouldBeNil)

		err = pt.createPrometheusFromSavedResource()
		So(err, ShouldBeNil)
	}

	pt.beforeBackup = false
	err := pt.waitForPodPrometheus(10 * time.Minute)
	So(err, ShouldBeNil)
	qresp := &queryResponse{}

	timeout := time.After(2 * time.Minute)
	tick := time.Tick(2 * time.Second)
	timedout := false
	done := false
	for {
		select {
		case <-timeout:
			timedout = true
			pt.log.Infof("Timedout: while hitting Prometheus API: %v", err)
		case <-tick:
			err = qresp.connectToPrometheusApi(pt.domain, pt.port, pt.api, pt.metricQuery, pt.pointInTime.formmattedValue)
			if err == nil {
				done = true
			}
		}

		if done || timedout {
			break
		}
	}

	So(err, ShouldBeNil)

	if len(qresp.Data.Result) > 0 && len(qresp.Data.Result[0].Value) > 0 {
		values := qresp.Data.Result[0].Value
		for _, something := range values {

			_, s, err := whatIsThisThing(something)
			So(err, ShouldBeNil)
			if s != "" {
				pt.finalResult = s
			}

		}

	}

	So(strings.TrimSpace(pt.finalResult), ShouldEqual, strings.TrimSpace(pt.expectedResult))
}

func (pt *prometheusTest) DeleteResources(namespace string) {
	// It needs to be implemented for this test.
	err := pt.waitForPodPrometheus(1 * time.Minute)
	So(err, ShouldBeNil)

	err = pt.deleteServices(prometheusNS, prometheusServiceName, prometheusLabelSelector)
	So(err, ShouldBeNil)

	err = pt.deletePrometheus(prometheusNS, prometheusName)
	So(err, ShouldBeNil)

	err = pt.deletePod(prometheusNS, prometheusPodName, prometheusLabelSelector)
	So(err, ShouldBeNil)

	err = pt.deletePVC(prometheusNS, prometheusPvcName, prometheusLabelSelector)
	So(err, ShouldBeNil)

}

func (pt *prometheusTest) waitForPodPrometheus(waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pod, err := pt.coreClient.CoreV1().Pods(prometheusNS).Get(prometheusPodName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			return fmt.Errorf("Pod did not start within given time  %v: %+v", waitmax, pod)
		case <-tick:
			pod, err := pt.coreClient.CoreV1().Pods(prometheusNS).Get(prometheusPodName, metav1.GetOptions{})
			pt.log.Info("Waiting for prometheus pod to be up!")
			if err != nil && strings.Contains(err.Error(), "not found") {
				// If Pod is not there, we hope the pod will come soon
				break
			} else if err != nil {
				return fmt.Errorf("Error in fetching Prometheus pod %s: ", err.Error())
			}

			// If Pod condition is not ready the for will continue until timeout
			if len(pod.Status.Conditions) > 0 {
				conditions := pod.Status.Conditions
				for _, cond := range conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						pt.log.Info("Prometheus pod is ready!")
						return nil
					}
				}
			}

			// Succeeded or Failed or Unknoen are taken as a error
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
				return fmt.Errorf("Prometheus in state %v: \n%+v", pod.Status.Phase, pod)
			}
		}
	}
}

func (pt *prometheusTest) deleteServices(namespace, serviceName, labelSelector string) error {
	deletePolicy := metav1.DeletePropagationForeground

	serviceList, err := pt.coreClient.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, service := range serviceList.Items {
		if service.Name == serviceName {
			err := pt.coreClient.CoreV1().Services(namespace).Delete(serviceName, &metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (pt *prometheusTest) savePrometheusResource(namespace, name string) error {
	backedupPrometheusLocal, err := pt.prometheusClient.MonitoringV1().Prometheuses(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	backedupPrometheus = backedupPrometheusLocal
	byteArr, err := json.Marshal(backedupPrometheus)
	if err != nil {
		return err
	}
	pt.log.Infoln("------------- Prometheus obj after saving -------------------")
	pt.log.Infof("%v", string(byteArr))
	return nil
}

func (pt *prometheusTest) createPrometheusFromSavedResource() error {
	backedupPrometheus.ObjectMeta.ResourceVersion = ""
	backedupPrometheus.Status = nil
	backedupPrometheus.Generation = 0
	prometheusAnnotations := backedupPrometheus.Annotations
	backedupPrometheus.ObjectMeta = metav1.ObjectMeta{
		Annotations:       prometheusAnnotations,
		Name:              backedupPrometheus.Name,
		Labels:            backedupPrometheus.Labels,
		Namespace:         backedupPrometheus.Namespace,
		CreationTimestamp: backedupPrometheus.CreationTimestamp,
	}
	podAnnotations := backedupPrometheus.Spec.PodMetadata.Annotations
	backedupPrometheus.Spec.PodMetadata = &metav1.ObjectMeta{
		Annotations:       podAnnotations,
		CreationTimestamp: backedupPrometheus.CreationTimestamp,
	}
	backedupPrometheus.Spec.Storage.VolumeClaimTemplate.ObjectMeta = metav1.ObjectMeta{
		CreationTimestamp: backedupPrometheus.CreationTimestamp,
	}

	byteArr, err := json.Marshal(backedupPrometheus)
	if err != nil {
		return err
	}

	pt.log.Infoln("------------- Prometheus obj while creation -------------------")
	pt.log.Infof("%v", string(byteArr))
	for {
		_, err := pt.prometheusClient.MonitoringV1().Prometheuses(backedupPrometheus.Namespace).Get(backedupPrometheus.Name, metav1.GetOptions{})
		if err != nil && strings.Contains(err.Error(), "not found") {
			log.Printf("err: %s", err.Error())
			break
		}
	}

	_, err = pt.prometheusClient.MonitoringV1().Prometheuses(backedupPrometheus.ObjectMeta.Namespace).Create(backedupPrometheus)
	return err
}

func (pt *prometheusTest) deletePrometheus(namespace, name string) error {
	pt.savePrometheusResource(namespace, name)
	deletePolicy := metav1.DeletePropagationForeground
	err := pt.prometheusClient.MonitoringV1().Prometheuses(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return err
	}

	pt.log.Info("Deletion of Prometheus: monitoring is successful!")
	return nil
}

func (pt *prometheusTest) deleteStatefulset(namespace, statefulsetName string) error {
	deletePolicy := metav1.DeletePropagationForeground

	collection := pt.coreClient.AppsV1().StatefulSets(namespace)
	err := collection.Delete(statefulsetName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return err
	}
	pt.log.Info("Deletion of sts: Prometheus is successful!")
	return nil
}

func (pt *prometheusTest) deletePod(namespace, podName, labelSelector string) error {
	podList, err := pt.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if pod.Name == podName {
			// Delete Pod
			err = pt.coreClient.CoreV1().Pods(namespace).Delete(podName, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (pt *prometheusTest) deletePVC(namespace, pvcName, labelSelector string) error {
	pvcList, err := pt.coreClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	for _, pvc := range pvcList.Items {
		if pvc.Name == pvcName {
			err = pt.coreClient.CoreV1().PersistentVolumeClaims(namespace).Delete(pvcName, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			pt.log.Infof("Deletion of pvc: %v is successful!", pvcName)
		}
	}
	return nil
}
