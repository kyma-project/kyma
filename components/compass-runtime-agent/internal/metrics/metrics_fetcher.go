package metrics

import (
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type MetricsFetcher interface {
	FetchNodeMetrics() ([]NodeMetrics, error)
}

type metricsFetcher struct {
	metricsClientSet v1beta1.NodeMetricsInterface
}

func newMetricsFetcher(c clientset.Interface) MetricsFetcher {
	return &metricsFetcher{
		metricsClientSet: c.MetricsV1beta1().NodeMetricses(),
	}
}

func (m *metricsFetcher) FetchNodeMetrics() ([]NodeMetrics, error) {
	metricList, err := m.metricsClientSet.List(meta.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list metrics")
	}

	clusterUsage := make([]NodeMetrics, 0)

	for _, metric := range metricList.Items {
		clusterUsage = append(clusterUsage, NodeMetrics{
			Name: metric.Name,
			Usage: ResourceGroup{
				CPU:              metric.Usage.Cpu().String(),
				EphemeralStorage: metric.Usage.StorageEphemeral().String(),
				Memory:           metric.Usage.Memory().String(),
				Pods:             metric.Usage.Pods().String(),
			},
			StartCollectingTimestamp: metric.Timestamp.Time,
		})
	}

	return clusterUsage, nil
}
