package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"kyma-project.io/compass-runtime-agent/internal/metrics/mocks"
)

const (
	loggingInterval = time.Millisecond
	loggingWaitTime = time.Millisecond * 10
)

type Log struct {
	Level       string      `json:"level"`
	Metrics     bool        `json:"metrics"`
	Msg         string      `json:"msg"`
	Time        time.Time   `json:"time"`
	ClusterInfo ClusterInfo `json:"clusterInfo"`
}

func Test_Log(t *testing.T) {
	t.Run("should log metrics", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset(
			&corev1.Node{
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
			},
			&corev1.PersistentVolume{
				ObjectMeta: meta.ObjectMeta{
					Name: "somename",
				},
				Spec: corev1.PersistentVolumeSpec{
					Capacity: corev1.ResourceList{
						corev1.ResourceStorage: *resource.NewQuantity(1, resource.BinarySI),
					},
					ClaimRef: &corev1.ObjectReference{
						Namespace: "claimnamespace",
						Name:      "claimname",
					},
				},
			},
		)

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
					corev1.ResourceEphemeralStorage: *resource.NewQuantity(0, resource.BinarySI),
					corev1.ResourcePods:             *resource.NewQuantity(0, resource.DecimalSI),
				},
				Timestamp: meta.Time{Time: time.Now()},
			}},
		}, nil)

		logger := NewMetricsLogger(resourcesClientset, metricsClientset, loggingInterval)

		quitChannel := make(chan struct{})
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go func() {
			err := logger.Start(quitChannel)
			assert.NoError(t, err, "failed to finish gracefully")
		}()

		time.Sleep(loggingWaitTime)
		quitChannel <- struct{}{}
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		logsSlice := strings.Split(logs, "\n")
		require.NotEqual(t, 0, len(logsSlice), "there are no logs")

		var singleLog Log
		err := json.Unmarshal([]byte(logsSlice[0]), &singleLog)
		require.NoError(t, err, "failed to unmarshal the first log")

		assert.Equal(t, true, singleLog.Metrics)
		assert.Equal(t, "info", singleLog.Level)
		assert.Equal(t, "Cluster metrics logged successfully.", singleLog.Msg)
		assert.NotEqual(t, 0, len(singleLog.ClusterInfo.Resources))
		assert.NotEqual(t, 0, len(singleLog.ClusterInfo.Usage))
		assert.NotEqual(t, 0, len(singleLog.ClusterInfo.Volumes))

		assert.Equal(t, true, strings.Contains(logs, "Logging stopped."), "did not finish gracefully")
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

		quitChannel := make(chan struct{}, 1)
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go func() {
			err := logger.Start(quitChannel)
			assert.NoError(t, err, "failed to finish gracefully")
		}()

		time.Sleep(loggingWaitTime)
		quitChannel <- struct{}{}
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		assert.Equal(t, true, strings.Contains(logs, "\"resources\":[]"), "resources are not empty array")
		assert.Equal(t, true, strings.Contains(logs, "\"usage\":[]"), "usage is not empty array")
		assert.Equal(t, true, strings.Contains(logs, "\"persistentVolumes\":[]"), "persistentVolumes is not empty array")
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

		quitChannel := make(chan struct{}, 1)
		defer close(quitChannel)

		var buffer bytes.Buffer
		log.SetOutput(&buffer)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		// when
		go func() {
			err := logger.Start(quitChannel)
			assert.NoError(t, err, "failed to finish gracefully")
		}()

		time.Sleep(loggingWaitTime)
		quitChannel <- struct{}{}
		time.Sleep(loggingWaitTime)

		// then
		logs := buffer.String()
		assert.Equal(t, true, strings.Contains(logs, "error"), "did not log an error")
		assert.Equal(t, false, strings.Contains(logs, "Cluster metrics logged successfully."), "did log metrics")
		assert.Equal(t, true, strings.Contains(logs, "Logging stopped."), "did not finish gracefully")
	})
}
