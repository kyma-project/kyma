/*
Copyright 2019 The Kyma Authors.

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

package utils

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// Semantic can do semantic deep equality checks for API objects.
// Fields which are not relevant for our reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	ksvcEqual,
)

// ksvcEqual asserts the equality of two Knative Service objects.
func ksvcEqual(s1, s2 *servingv1.Service) bool {
	if s1 == s2 {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}

	if !reflect.DeepEqual(s1.Labels, s2.Labels) {
		return false
	}

	cst1 := &s1.Spec.ConfigurationSpec.Template
	cst2 := &s2.Spec.ConfigurationSpec.Template
	if cst1 == nil && cst2 != nil {
		return false
	}
	if cst1 != nil {
		if cst2 == nil {
			return false
		}

		ps1 := &cst1.Spec.PodSpec
		ps2 := &cst2.Spec.PodSpec
		if !podSpecEqual(ps1, ps2) {
			return false
		}
	}

	return true
}

// podSpecEqual asserts the equality of two PodSpec objects.
func podSpecEqual(ps1, ps2 *corev1.PodSpec) bool {
	if ps1 == ps2 {
		return true
	}
	if ps1 == nil || ps2 == nil {
		return false
	}

	cs1, cs2 := ps1.Containers, ps2.Containers
	if len(cs1) != len(cs2) {
		return false
	}
	for i := range cs1 {
		if !containerEqual(&cs1[i], &cs2[i]) {
			return false
		}
	}

	return ps1.ServiceAccountName == ps2.ServiceAccountName
}

// containerEqual asserts the equality of two Container objects.
func containerEqual(c1, c2 *corev1.Container) bool {
	if c1.Image != c2.Image {
		return false
	}

	if !resourceListEqual(c1.Resources.Requests, c2.Resources.Requests) {
		return false
	}

	ps1, ps2 := c1.Ports, c2.Ports
	if len(ps1) != len(ps2) {
		return false
	}
	for i := range ps1 {
		p1, p2 := &ps1[i], &ps2[i]

		if p1.Name != p2.Name ||
			p1.ContainerPort != p2.ContainerPort ||
			realProto(p1.Protocol) != realProto(p2.Protocol) {
			return false
		}
	}

	return reflect.DeepEqual(c1.Env, c2.Env)
}

// resourceListEqual asserts the equality of two ResourceList objects.
func resourceListEqual(rl1, rl2 corev1.ResourceList) bool {
	for resName, q1 := range rl1 {
		q2, ok := rl2[resName]
		if !ok {
			return false
		}
		if q1.Cmp(q2) != 0 {
			return false
		}
	}

	return true
}

// default Protocol is TCP, so we assume empty equals TCP
// https://godoc.org/k8s.io/api/core/v1#ServicePort
func realProto(pr corev1.Protocol) corev1.Protocol {
	if pr == "" {
		return corev1.ProtocolTCP
	}
	return pr
}
