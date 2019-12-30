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

package object

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// Semantic can do semantic deep equality checks for API objects. Fields which
// are not relevant for the reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	channelEqual,
	ksvcEqual,
)

// channelEqual asserts the equality of two Channel objects.
func channelEqual(c1, c2 *messagingv1alpha1.Channel) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}

	if !reflect.DeepEqual(c1.Labels, c2.Labels) {
		return false
	}
	if !reflect.DeepEqual(c1.Annotations, c2.Annotations) {
		return false
	}

	return true
}

// ksvcEqual asserts the equality of two Knative Service objects.
func ksvcEqual(s1, s2 *servingv1alpha1.Service) bool {
	if s1 == s2 {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}

	if !reflect.DeepEqual(s1.Labels, s2.Labels) {
		return false
	}
	if !reflect.DeepEqual(s1.Annotations, s2.Annotations) {
		return false
	}

	cst1 := s1.Spec.ConfigurationSpec.Template
	cst2 := s2.Spec.ConfigurationSpec.Template
	if cst1 == nil && cst2 != nil {
		return false
	}
	if cst1 != nil {
		if cst2 == nil {
			return false
		}

		if !reflect.DeepEqual(cst1.Annotations, cst2.Annotations) {
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

	if ps1.ServiceAccountName != ps2.ServiceAccountName {
		return false
	}

	return true
}

// containerEqual asserts the equality of two Container objects.
func containerEqual(c1, c2 *corev1.Container) bool {
	if c1.Image != c2.Image {
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

	if !reflect.DeepEqual(c1.Env, c2.Env) {
		return false
	}

	if !probeEqual(c1.ReadinessProbe, c2.ReadinessProbe) {
		return false
	}

	return true
}

// probeEqual asserts the equality of two Probe objects.
func probeEqual(p1, p2 *corev1.Probe) bool {
	if p1 == p2 {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}

	if p1.InitialDelaySeconds != p2.InitialDelaySeconds ||
		p1.TimeoutSeconds != p2.TimeoutSeconds ||
		p1.PeriodSeconds != p2.PeriodSeconds ||
		// Knative sets a default when that value is 0
		p1.SuccessThreshold != p2.SuccessThreshold && !(p1.SuccessThreshold == 0 || p2.SuccessThreshold == 0) ||
		p1.FailureThreshold != p2.FailureThreshold {

		return false
	}

	if !handlerEqual(&p1.Handler, &p2.Handler) {
		return false
	}

	return true
}

// handlerEqual asserts the equality of two Handler objects.
func handlerEqual(h1, h2 *corev1.Handler) bool {
	if h1 == h2 {
		return true
	}
	if h1 == nil || h2 == nil {
		return false
	}

	hg1, hg2 := h1.HTTPGet, h2.HTTPGet
	if hg1 == nil && hg2 != nil {
		return false
	}
	if hg1 != nil {
		if hg2 == nil {
			return false
		}

		if hg1.Path != hg2.Path {
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
