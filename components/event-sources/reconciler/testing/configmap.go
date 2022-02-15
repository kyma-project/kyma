package testing

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConfigMap creates a ConfigMap object.
func NewConfigMap(ns, name string, opts ...ConfigMapOption) *corev1.ConfigMap {
	cmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}

	for _, opt := range opts {
		opt(cmap)
	}

	return cmap
}

// ConfigMapOption is a functional option for ConfigMap objects.
type ConfigMapOption func(*corev1.ConfigMap)

// WithData sets the data of a ConfigMap.
func WithData(data map[string]string) ConfigMapOption {
	return func(cmap *corev1.ConfigMap) {
		cmap.Data = data
	}
}
