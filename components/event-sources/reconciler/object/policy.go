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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	authenticationv1alpha1api "istio.io/api/authentication/v1alpha1"
	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
)

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

// ApplyExistingPolicyAttributes copies some important attributes from a given
// source Policy to a destination Policy.
func ApplyExistingPolicyAttributes(src, dst *authenticationv1alpha1.Policy) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion
}

// WithTarget sets the target name of the Policy for a Knative Service which
// has metrics end-points
func WithTarget(target string) ObjectOption {
	return func(o metav1.Object) {
		p := o.(*authenticationv1alpha1.Policy)
		p.Spec = authenticationv1alpha1api.Policy{
			Targets: []*authenticationv1alpha1api.TargetSelector{
				{Name: target},
			},
		}
	}
}
