package istio

import (
	securityv1beta1apis "istio.io/api/security/v1beta1"
	istiov1beta1apis "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectOption is a functional option for API objects builders.
type ObjectOption func(metav1.Object)

// NewPeerAuthentication returns a new PeerAuthentication with empty Spec
func NewPeerAuthentication(ns, name string, opts ...ObjectOption) *securityv1beta1.PeerAuthentication {
	s := &securityv1beta1.PeerAuthentication{
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

// WithPermissiveMode sets the mTLS mode of the PeerAuthentication to Permissive for the given port
func WithPermissiveMode(port uint32) ObjectOption {
	return func(o metav1.Object) {
		p := o.(*securityv1beta1.PeerAuthentication)
		p.Spec.PortLevelMtls = map[uint32]*securityv1beta1apis.PeerAuthentication_MutualTLS{
			port: {
				Mode: securityv1beta1apis.PeerAuthentication_MutualTLS_PERMISSIVE,
			},
		}
	}
}

// WithSelectorSpec selects a workload based on labels
func WithSelectorSpec(labels map[string]string) ObjectOption {
	return func(o metav1.Object) {
		p := o.(*securityv1beta1.PeerAuthentication)
		p.Spec.Selector = &istiov1beta1apis.WorkloadSelector{
			MatchLabels: labels,
		}
	}
}

// WithLabel sets the value of an API object's label.
func WithLabel(key, val string) ObjectOption {
	return func(o metav1.Object) {
		lbls := o.GetLabels()
		if lbls == nil {
			lbls = make(map[string]string, 1)
			o.SetLabels(lbls)
		}
		lbls[key] = val
	}
}
