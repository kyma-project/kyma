package metrics

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"kyma-project.io/compass-runtime-agent/internal/metrics/mocks"
)

func Test_FetchNodeMetrics(t *testing.T) {
	t.Run("should fetch nodes metrics", func(t *testing.T) {
		// given
		now := time.Now()
		metricsClientset := &mocks.MetricsClientsetInterface{}
		metricsV1beta1 := &mocks.MetricsV1beta1Interface{}
		nodeMetrics := &mocks.NodeMetricsInterface{}
		metricsClientset.On("MetricsV1beta1").Return(metricsV1beta1)
		metricsV1beta1.On("NodeMetricses").Return(nodeMetrics)
		nodeMetrics.On("List", v1.ListOptions{}).Return(&v1beta1.NodeMetricsList{
			Items: []v1beta1.NodeMetrics{{
				ObjectMeta: v1.ObjectMeta{
					Name: "somename",
				},
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(1, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourceEphemeralStorage: *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourcePods:             *resource.NewQuantity(1, resource.DecimalSI),
				},
				Timestamp: v1.Time{Time: now},
			}},
		}, nil)

		metricsFetcher := newMetricsFetcher(metricsClientset)

		// when
		metrics, err := metricsFetcher.FetchNodeMetrics()
		require.NoError(t, err)

		// then
		require.Equal(t, 1, len(metrics))
		assert.Equal(t, "somename", metrics[0].Name)
		assert.Equal(t, "1", metrics[0].Usage.CPU)
		assert.Equal(t, "1", metrics[0].Usage.Memory)
		assert.Equal(t, "1", metrics[0].Usage.EphemeralStorage)
		assert.Equal(t, "1", metrics[0].Usage.Pods)
		assert.Equal(t, now, metrics[0].StartCollectingTimestamp)
	})

	t.Run("should not fail if no node metrics", func(t *testing.T) {
		// given
		metricsClientset := &mocks.MetricsClientsetInterface{}
		metricsV1beta1 := &mocks.MetricsV1beta1Interface{}
		nodeMetrics := &mocks.NodeMetricsInterface{}
		metricsClientset.On("MetricsV1beta1").Return(metricsV1beta1)
		metricsV1beta1.On("NodeMetricses").Return(nodeMetrics)
		nodeMetrics.On("List", v1.ListOptions{}).Return(&v1beta1.NodeMetricsList{}, nil)

		metricsFetcher := newMetricsFetcher(metricsClientset)

		// when
		metrics, err := metricsFetcher.FetchNodeMetrics()
		require.NoError(t, err)

		// then
		assert.Equal(t, 0, len(metrics))
	})

	t.Run("should fail if list failed", func(t *testing.T) {
		// given
		metricsClientset := &mocks.MetricsClientsetInterface{}
		metricsV1beta1 := &mocks.MetricsV1beta1Interface{}
		nodeMetrics := &mocks.NodeMetricsInterface{}
		metricsClientset.On("MetricsV1beta1").Return(metricsV1beta1)
		metricsV1beta1.On("NodeMetricses").Return(nodeMetrics)
		nodeMetrics.On("List", v1.ListOptions{}).Return(nil, fmt.Errorf("someerror"))

		metricsFetcher := newMetricsFetcher(metricsClientset)

		// when
		_, err := metricsFetcher.FetchNodeMetrics()

		// then
		assert.Error(t, err)
	})
}
