package metrics

import (
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

//go:generate mockery --name=MetricsClientsetInterface
type MetricsClientsetInterface interface {
	clientset.Interface
}

//go:generate mockery --name=MetricsV1beta1Interface
type MetricsV1beta1Interface interface {
	v1beta1.MetricsV1beta1Interface
}

//go:generate mockery --name=NodeMetricsInterface
type NodeMetricsInterface interface {
	v1beta1.NodeMetricsInterface
}
