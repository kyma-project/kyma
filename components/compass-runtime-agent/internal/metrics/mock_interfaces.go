package metrics

import (
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type MetricsClientsetInterface interface {
	clientset.Interface
}

type MetricsV1beta1Interface interface {
	v1beta1.MetricsV1beta1Interface
}

type NodeMetricsInterface interface {
	v1beta1.NodeMetricsInterface
}
