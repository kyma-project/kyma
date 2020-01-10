package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"

	prometheus "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MetricsUpgradeTest compares metrics from before and after upgrade
type MetricsUpgradeTest struct {
	k8sCli        kubernetes.Interface
	namespace     string
	prometheusAPI v1.API
	log           logrus.FieldLogger
}

// NewMetricsUpgradeTest returns a new instance of the MetricsUpgradeTest
func NewMetricsUpgradeTest(k8sCli kubernetes.Interface) (*MetricsUpgradeTest, error) {
	client, err := prometheus.NewClient(prometheus.Config{Address: fmt.Sprintf("%v:%v", prometheusDomain, prometheusPort)})
	if err != nil {
		return nil, err
	}
	promAPI := v1.NewAPI(client)

	return &MetricsUpgradeTest{
		prometheusAPI: promAPI,
		k8sCli:        k8sCli,
	}, nil

}

// CreateResources retrieves metrics and stores the value in an configmap
func (ut *MetricsUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log

	time := time.Now().Add(-time.Minute)
	result, err := ut.collectMetrics(time)
	if err != nil {
		return err
	}
	err = ut.storeMetrics(result, time)
	return err
}

// TestResources retrieves previously installed metrics values and compares it to the metrics returned for the same query and the same time
func (ut *MetricsUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log
	return ut.compareMetrics()
}

const (
	prometheusDomain     = "http://monitoring-prometheus.kyma-system"
	prometheusNS         = "kyma-system"
	prometheusPort       = "9090"
	metricsName          = "kube_pod_container_resource_requests_cpu_cores"
	metricsQuery         = "max(sum(kube_pod_container_resource_requests_cpu_cores) by (instance))"
	metricsConfigMapName = "metrics-upgrade-test"
	metricsDataField     = "response"
	metricsTimeField     = "collected_at"
)

func (ut *MetricsUpgradeTest) collectMetrics(time time.Time) (model.Vector, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, _, err := ut.prometheusAPI.Query(ctx, metricsQuery, time)
	if err != nil {
		return nil, err
	}
	if result.Type() == model.ValVector {
		ut.log.Debugln(result.(model.Vector))
		return result.(model.Vector), nil
	}
	return nil, fmt.Errorf("%v should be of type vector but is of type %t", result, result)
}

func (ut *MetricsUpgradeTest) storeMetrics(value model.Vector, time time.Time) error {
	promValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	timeValue, err := json.Marshal(time)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: metricsConfigMapName,
		},
		Data: map[string]string{
			metricsDataField: string(promValue),
			metricsTimeField: string(timeValue),
		},
	}
	_, err = ut.k8sCli.CoreV1().ConfigMaps(ut.namespace).Create(cm)
	return err
}

func (ut *MetricsUpgradeTest) retrievePreviousMetrics() (time.Time, model.Vector, error) {

	cm, err := ut.k8sCli.CoreV1().ConfigMaps(ut.namespace).Get(metricsConfigMapName, metav1.GetOptions{})
	if err != nil {
		return time.Time{}, nil, err
	}

	value := model.Vector{}
	err = json.Unmarshal([]byte(cm.Data[metricsDataField]), &value)
	if err != nil {
		return time.Time{}, nil, err
	}

	timeValue := time.Time{}
	err = json.Unmarshal([]byte(cm.Data[metricsTimeField]), &timeValue)
	if err != nil {
		return time.Time{}, nil, err
	}
	return timeValue, value, nil
}

func (ut *MetricsUpgradeTest) compareMetrics() error {
	time, previous, err := ut.retrievePreviousMetrics()
	if err != nil {
		return err
	}
	current, err := ut.collectMetrics(time)
	if err != nil {
		return err
	}
	ut.log.Debugln(previous)
	ut.log.Debugln(current)
	if !cmp.Equal(previous, current) {
		return fmt.Errorf("retrieved data not equal: before: %+v, after: %+v", previous, current)
	}
	return nil
}
