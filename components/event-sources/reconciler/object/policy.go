package object

import (
	authenticationv1alpha1api "istio.io/api/authentication/v1alpha1"
	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const targetPort = "http-usermetric"

// NewPolicy creates a Policy object.
func NewPolicy(ns, name string, opts ...ObjectOption) *authenticationv1alpha1.Policy {
	s := &authenticationv1alpha1.Policy{
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

// ApplyExistingPeerAuthenticationAttributes copies some important attributes from a given
// source Policy to a destination Policy.
func ApplyExistingPeerAuthenticationAttributes(src, dst *securityv1beta1.PeerAuthentication) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion
}

// WithTargetDeprecated sets the target name of the Policy for a Deployment which
// has metrics end-point
func WithTargetDeprecated(target string) ObjectOption {
	return func(o metav1.Object) {
		p := o.(*authenticationv1alpha1.Policy)
		p.Spec.Targets = []*authenticationv1alpha1api.TargetSelector{
			{
				Name: target,
				Ports: []*authenticationv1alpha1api.PortSelector{
					{
						Port: &authenticationv1alpha1api.PortSelector_Name{
							Name: targetPort,
						},
					},
				},
			},
		}
	}
}

// WithPermissiveModeDeprecated sets the mTLS mode of the policy to Permissive
func WithPermissiveModeDeprecated() ObjectOption {
	return func(o metav1.Object) {
		p := o.(*authenticationv1alpha1.Policy)
		p.Spec.Peers = []*authenticationv1alpha1api.PeerAuthenticationMethod{
			{
				Params: &authenticationv1alpha1api.PeerAuthenticationMethod_Mtls{
					Mtls: &authenticationv1alpha1api.MutualTls{
						Mode: authenticationv1alpha1api.MutualTls_PERMISSIVE,
					},
				},
			},
		}
	}
}
