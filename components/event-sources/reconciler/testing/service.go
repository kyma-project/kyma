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

package testing

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	routeconfig "knative.dev/serving/pkg/reconciler/route/config"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
)

const DefaultHTTPProbePath = "/healthz"

// NewService creates a Service object.
func NewService(ns, name string, opts ...ServiceOption) *servingv1alpha1.Service {
	s := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			// assume all Services are created with "cluster-local" visibility
			Labels: map[string]string{
				routeconfig.VisibilityLabelKey: routeconfig.VisibilityClusterLocal,
			},
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	// assume all Services expose a "/healthz" endpoint for probes
	if s.Spec.ConfigurationSpec.Template != nil &&
		s.Spec.ConfigurationSpec.Template.Spec.Containers != nil {

		s.Spec.ConfigurationSpec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: DefaultHTTPProbePath,
				},
			},
		}
	}

	return s
}

// ServiceOption is a functional option for Service objects.
type ServiceOption func(*servingv1alpha1.Service)

// WithServiceReady marks the Service as Ready.
func WithServiceReady(s *servingv1alpha1.Service) {
	s.Status.SetConditions(apis.Conditions{{
		Type:   apis.ConditionReady,
		Status: corev1.ConditionTrue,
	}})
}

// WithServiceNotReady marks the Service as not Ready.
func WithServiceNotReady(s *servingv1alpha1.Service) {
	s.Status.SetConditions(apis.Conditions{{
		Type:   apis.ConditionReady,
		Status: corev1.ConditionFalse,
	}})
}

// WithServiceController sets the controller of a Service.
func WithServiceController(srcName string) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		gvk := sourcesv1alpha1.HTTPSourceGVK()

		s.OwnerReferences = []metav1.OwnerReference{{
			APIVersion:         gvk.GroupVersion().String(),
			Kind:               gvk.Kind,
			Name:               srcName,
			UID:                uid,
			Controller:         ptr.Bool(true),
			BlockOwnerDeletion: ptr.Bool(true),
		}}
	}
}

// WithServiceContainer configures a container for a Service.
func WithServiceContainer(img string, port int32, ev []corev1.EnvVar) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		s.Spec.ConfigurationSpec.Template = &servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				RevisionSpec: servingv1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: img,
							Ports: []corev1.ContainerPort{{
								ContainerPort: port,
							}},
							Env: ev,
						}},
					},
				},
			},
		}
	}
}
