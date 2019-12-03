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

package objects

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	"knative.dev/serving/pkg/apis/serving"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// List of annotations set on Knative Serving objects by the Knative Serving
// admission webhook.
var knativeServingAnnotations = []string{
	serving.GroupName + apis.CreatorAnnotationSuffix,
	serving.GroupName + apis.UpdaterAnnotationSuffix,
}

// NewService creates a Service object.
func NewService(ns, name string, opts ...ServiceOption) *servingv1alpha1.Service {
	s := &servingv1alpha1.Service{
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

// ServiceOption is a functional option for Service objects.
type ServiceOption func(*servingv1alpha1.Service)

// WithServiceControllerRef sets the controller reference of a Service.
func WithServiceControllerRef(or *metav1.OwnerReference) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		s.OwnerReferences = []metav1.OwnerReference{*or}
	}
}

// WithServiceLabel sets the value of a Service label.
func WithServiceLabel(key, val string) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		if s.Labels == nil {
			s.Labels = make(map[string]string, 1)
		}
		s.Labels[key] = val
	}
}

// WithContainerImage sets the container image of a Service.
func WithContainerImage(img string) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		firstServiceContainer(s).Image = img
	}
}

// WithContainerPort sets the container port of a Service.
func WithContainerPort(port int32) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		ports := &firstServiceContainer(s).Ports

		*ports = append(*ports, corev1.ContainerPort{
			// empty name defaults to http/1.1 protocol
			ContainerPort: port,
		})
	}
}

// WithContainerEnvVar sets the value of a container env var.
func WithContainerEnvVar(name, val string) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		envvars := &firstServiceContainer(s).Env

		*envvars = append(*envvars, corev1.EnvVar{
			Name:  name,
			Value: val,
		})
	}
}

// WithContainerProbe sets the HTTP readiness probe of a container.
func WithContainerProbe(path string) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		firstServiceContainer(s).ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: path,
					// setting port explicitely is illegal in a Knative Service
				},
			},
		}
	}
}

// firstServiceContainer returns the first Container definition of a Service. A
// new empty Container is injected if the Service does not contain any.
func firstServiceContainer(s *servingv1alpha1.Service) *corev1.Container {
	if s.Spec.ConfigurationSpec.Template == nil {
		s.Spec.ConfigurationSpec.Template = &servingv1alpha1.RevisionTemplateSpec{}
	}

	containers := &s.Spec.ConfigurationSpec.Template.Spec.Containers
	if *containers == nil {
		*containers = make([]corev1.Container, 1)
	}
	return &(*containers)[0]
}

// ApplyExistingServiceAttributes copies some important attributes from a given
// source Service to a destination Service.
func ApplyExistingServiceAttributes(src, dst *servingv1alpha1.Service) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion

	// immutable Knative annotations must be preserved
	for _, ann := range knativeServingAnnotations {
		if val, ok := src.Annotations[ann]; ok {
			metav1.SetMetaDataAnnotation(&dst.ObjectMeta, ann, val)
		}
	}

	// preserve status to avoid resetting conditions
	dst.Status = src.Status
}
