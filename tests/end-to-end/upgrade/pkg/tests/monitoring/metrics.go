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

type metricsUpgradeTest struct {
	k8sCli        kubernetes.Interface
	namespace     string
	prometheusApi v1.API
	log           logrus.FieldLogger
}

func NewMetricsUpgradeTest(k8sCli kubernetes.Interface) (*metricsUpgradeTest, error) {
	client, err := prometheus.NewClient(prometheus.Config{Address: fmt.Sprintf("%v:%v", domain, port)})
	if err != nil {
		return nil, err
	}
	promApi := v1.NewAPI(client)

	return &metricsUpgradeTest{
		prometheusApi: promApi,
		k8sCli:        k8sCli,
	}, nil

}

func (ut *metricsUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log

	time := time.Now().Add(-time.Minute)
	log.Debugln("before collect")
	result, err := ut.collectMetrics(time)
	if err != nil {
		return err
	}
	log.Debugln("after collect")
	log.Debugln("before store")
	err = ut.storeMetrics(result, time)
	log.Debugln("after store")
	return err
}

func (ut *metricsUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ut.namespace = namespace
	ut.log = log
	return ut.compareMetrics()
}

const (
	domain        = "http://monitoring-prometheus.kyma-system"
	prometheusNS  = "kyma-system"
	metricsQuery  = "max(sum(kube_pod_container_resource_requests_cpu_cores) by (instance))"
	port          = "9090"
	metricName    = "kube_pod_container_resource_requests_cpu_cores"
	configMapName = "metrics-upgrade-test"
	dataField     = "response"
	timeField     = "collected_at"
)

func (ut *metricsUpgradeTest) collectMetrics(time time.Time) (model.Vector, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := ut.prometheusApi.Query(ctx, metricsQuery, time)
	if err != nil {
		return nil, err
	}
	if result.Type() == model.ValVector {
		ut.log.Debugln(result.(model.Vector))
		return result.(model.Vector), nil
	}
	return nil, fmt.Errorf("%v should be of type vector but is of type %t", result, result)
}

func (ut *metricsUpgradeTest) storeMetrics(value model.Vector, time time.Time) error {
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
			Name: configMapName,
		},
		Data: map[string]string{
			dataField: string(promValue),
			timeField: string(timeValue),
		},
	}
	_, err = ut.k8sCli.CoreV1().ConfigMaps(ut.namespace).Create(cm)
	return err
}

func (ut *metricsUpgradeTest) retrievePreviousMetrics() (time.Time, model.Vector, error) {

	cm, err := ut.k8sCli.CoreV1().ConfigMaps(ut.namespace).Get(configMapName, metav1.GetOptions{})
	if err != nil {
		return time.Time{}, nil, err
	}

	value := model.Vector{}
	err = json.Unmarshal([]byte(cm.Data[dataField]), &value)
	if err != nil {
		return time.Time{}, nil, err
	}

	timeValue := time.Time{}
	err = json.Unmarshal([]byte(cm.Data[timeField]), &timeValue)
	if err != nil {
		return time.Time{}, nil, err
	}
	return timeValue, value, nil
}

func (ut *metricsUpgradeTest) compareMetrics() error {
	time, previous, err := ut.retrievePreviousMetrics()
	if err != nil {
		return err
	}
	current, err := ut.collectMetrics(time)
	ut.log.Debugln(previous)
	ut.log.Debugln(current)
	if !cmp.Equal(previous, current) {
		return fmt.Errorf("retrieved data not equal: before: %+v, after: %+v", previous, current)
	}
	return nil
}
