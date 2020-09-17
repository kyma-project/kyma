package object

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewService creates a Service object.
func NewService(ns, name string, opts ...ObjectOption) *corev1.Service {
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithSelector(key, value string) ObjectOption {
	return func(o metav1.Object) {
		s := o.(*corev1.Service)
		s.Spec.Selector = map[string]string{key: value}
	}
}

func WithServicePort(name string, port int, containerPort int) ObjectOption {
	return func(o metav1.Object) {
		s := o.(*corev1.Service)
		s.Spec.Ports = append(s.Spec.Ports, corev1.ServicePort{
			Name:       name,
			Port:       int32(port),
			TargetPort: intstr.FromInt(containerPort),
		})
	}
}

// ApplyExistingServiceAttributes copies some important attributes from a given
// source Service to a destination Service.
func ApplyExistingServiceAttributes(src, dst *corev1.Service) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion
	dst.Spec.ClusterIP = src.Spec.ClusterIP
}
