package gateway

import (
	"time"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewStatusWatcher(gatewayServiceLister gatewayServiceLister, httpTimeout time.Duration) *gatewayStatusWatcher {
	return newStatusWatcher(gatewayServiceLister, httpTimeout)
}

func NewProvider(corev1Interface corev1.CoreV1Interface, integrationNamespace string, informerResyncPeriod time.Duration) *provider {
	return newProvider(corev1Interface, integrationNamespace, informerResyncPeriod)
}
