package metrics

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"kyma-project.io/compass-runtime-agent/internal/metrics/mocks"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	loggingInterval = time.Millisecond
	loggingWaitTime = time.Millisecond * 10
)

func Test_Log(t *testing.T) {
	t.Run("should log metrics", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset(&corev1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:   "somename",
				Labels: map[string]string{"beta.kubernetes.io/instance-type": "somelabel"},
			},
			Status: corev1.NodeStatus{
				Capacity: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(1, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourceEphemeralStorage: *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourcePods:             *resource.NewQuantity(1, resource.DecimalSI),
				},
			},
		})

		metricsClientset := &mocks.MetricsClientsetInterface{}
		metricsV1beta1 := &mocks.MetricsV1beta1Interface{}
		nodeMetrics := &mocks.NodeMetricsInterface{}
		metricsClientset.On("MetricsV1beta1").Return(metricsV1beta1)
		metricsV1beta1.On("NodeMetricses").Return(nodeMetrics)
		nodeMetrics.On("List", meta.ListOptions{}).Return(&v1beta1.NodeMetricsList{
			Items: []v1beta1.NodeMetrics{{
				ObjectMeta: meta.ObjectMeta{
					Name: "somename",
				},
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(1, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourceEphemeralStorage: *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourcePods:             *resource.NewQuantity(1, resource.DecimalSI),
				},
				Timestamp: meta.Time{Time: time.Now()},
			}},
		}, nil)

		logger := NewMetricsLogger(resourcesClientset, metricsClientset, loggingInterval)

		quitChannel := make(chan bool, 1)
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go logger.Log(quitChannel)

		time.Sleep(loggingWaitTime)
		quitChannel <- true
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		assert.Equal(t, true, strings.Contains(logs, "Cluster metrics logged successfully."), "did not log metrics")
		assert.Equal(t, true, strings.Contains(logs, "Logging stopped."), "did not finish gracefully")
		assert.Equal(t, true, strings.Contains(logs, "\"metrics\":true"), "metrics flag is not true")
		assert.Equal(t, true, strings.Contains(logs, "\"time\""), "there is no timestamp in the log")
		assert.Equal(t, false, strings.Contains(logs, "\"resources\":[]"), "resources are not empty")
		assert.Equal(t, false, strings.Contains(logs, "\"usage\":[]"), "usage is not empty")
		assert.Equal(t, false, strings.Contains(logs, "error"), "logged an error")
	})

	t.Run("should represent empty array as [], not null", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset()
		metricsClientset := &mocks.MetricsClientsetInterface{}
		metricsV1beta1 := &mocks.MetricsV1beta1Interface{}
		nodeMetrics := &mocks.NodeMetricsInterface{}
		metricsClientset.On("MetricsV1beta1").Return(metricsV1beta1)
		metricsV1beta1.On("NodeMetricses").Return(nodeMetrics)
		nodeMetrics.On("List", meta.ListOptions{}).Return(&v1beta1.NodeMetricsList{}, nil)

		logger := NewMetricsLogger(resourcesClientset, metricsClientset, loggingInterval)

		quitChannel := make(chan bool, 1)
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go logger.Log(quitChannel)

		time.Sleep(loggingWaitTime)
		quitChannel <- true
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		assert.Equal(t, true, strings.Contains(logs, "\"resources\":[]"), "resources are not empty array")
		assert.Equal(t, true, strings.Contains(logs, "\"usage\":[]"), "usage is not empty array")
		assert.Equal(t, false, strings.Contains(logs, "error"), "logged an error")
	})

	t.Run("should log error if occurred", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset()
		metricsClientset := &mocks.MetricsClientsetInterface{}
		metricsV1beta1 := &mocks.MetricsV1beta1Interface{}
		nodeMetrics := &mocks.NodeMetricsInterface{}
		metricsClientset.On("MetricsV1beta1").Return(metricsV1beta1)
		metricsV1beta1.On("NodeMetricses").Return(nodeMetrics)
		nodeMetrics.On("List", meta.ListOptions{}).Return(nil, fmt.Errorf("someerror"))

		logger := NewMetricsLogger(resourcesClientset, metricsClientset, loggingInterval)

		quitChannel := make(chan bool, 1)
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go logger.Log(quitChannel)

		time.Sleep(loggingWaitTime)
		quitChannel <- true
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		assert.Equal(t, true, strings.Contains(logs, "error"), "did not log an error")
		assert.Equal(t, false, strings.Contains(logs, "Cluster metrics logged successfully."), "did log metrics")
		assert.Equal(t, true, strings.Contains(logs, "Logging stopped."), "did not finish gracefully")
	})
}
