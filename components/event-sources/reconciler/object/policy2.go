/*
Copyright 2020 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package object

import (
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const targetPort2 = "http-usermetric"

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

func NewAuthorizationPolicy(ns, name string, opts ...ObjectOption) *securityv1beta1.AuthorizationPolicy {
	s := &securityv1beta1.AuthorizationPolicy{
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

func ApplyExistingPeerAuthentication(src, dst *securityv1beta1.PeerAuthentication) {
	// resourceVersion must be returned to the API server nmodified for optimistic concurrency,
	// as per Kubernetes API conventions
	dst.ResourceVersion = src.ResourceVersion
}

func ApplyExistingAuthorizationPolicy(src, dst *securityv1beta1.AuthorizationPolicy) {
	// resourceVersion must be returned to the API server nmodified for optimistic concurrency,
	// as per Kubernetes API conventions
	dst.ResourceVersion = src.ResourceVersion
}

// WithTargetAuthorizationPolicy sets the target name of the AuthorizationPolicy for a Knative Service which
// has metrics end-point
func WithTargetAuthorizationPolicy(target string) ObjectOption {
	return func(o metav1.Object) {
		p := o.(*securityv1beta1.AuthorizationPolicy)
		p.Spec.Selector.MatchLabels = map[string]string{"app": "knsvc", "version": "v1"}    // TODO find the real label for a ksvc
		p.Spec.Action = v1beta1.AuthorizationPolicy_ALLOW
		p.Spec.Rules = []*v1beta1.Rule {
			{
				From: []*v1beta1.Rule_From{
					{
						Source: &v1beta1.Source {
							Principals: []string{"*"},
						},
					},
				},
			To: []*v1beta1.Rule_To {
					{
						Operation: &v1beta1.Operation {
							Methods: []string{"GET"},
							Ports: []string{targetPort2},
						},
					},
				},
			},
		}
	}
}

// WithPermissiveMode2 sets the mTLS mode of the PeerAuthentication to Permissive
func WithPermissiveModePeerAuthentication() ObjectOption {
	return func(o metav1.Object) {
		pa := o.(*securityv1beta1.PeerAuthentication)
		pa.Spec.Selector.MatchLabels = map[string]string{"app": "knsvc", "version": "v1"}    // TODO find the real label for a ksvc
		pa.Spec.Mtls.Mode = v1beta1.PeerAuthentication_MutualTLS_PERMISSIVE
	}
}
