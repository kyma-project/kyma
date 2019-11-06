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
		svcOwnerRefs := &s.OwnerReferences

		switch {
		case *svcOwnerRefs == nil:
			*svcOwnerRefs = []metav1.OwnerReference{*or}
		case metav1.GetControllerOf(s) == nil:
			*svcOwnerRefs = append(*svcOwnerRefs, *or)
		default:
			for i, r := range *svcOwnerRefs {
				if r.Controller != nil && *r.Controller {
					(*svcOwnerRefs)[i] = *or
				}
			}
		}
	}
}

// WithContainerImage sets the container image of a Service. A minimal
// Container definition is injected if the Service does not contain any.
func WithContainerImage(img string) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		if s.Spec.ConfigurationSpec.Template == nil {
			s.Spec.ConfigurationSpec.Template = &servingv1alpha1.RevisionTemplateSpec{}
		}

		containers := &s.Spec.ConfigurationSpec.Template.Spec.Containers
		if *containers == nil {
			*containers = make([]corev1.Container, 1)
		}
		(*containers)[0].Image = img
	}
}

// WithExistingService copies some important attributes from an existing
// Service.
func WithExistingService(ksvc *servingv1alpha1.Service) ServiceOption {
	return func(s *servingv1alpha1.Service) {
		if ksvc == nil {
			return
		}

		// resourceVersion must be returned to the API server
		// unmodified for optimistic concurrency, as per Kubernetes API
		// conventions
		s.ResourceVersion = ksvc.ResourceVersion

		// immutable Knative annotations must be preserved
		for _, ann := range knativeServingAnnotations {
			if val, ok := ksvc.Annotations[ann]; ok {
				metav1.SetMetaDataAnnotation(&s.ObjectMeta, ann, val)
			}
		}

		// preserve owner references
		if ksvc.OwnerReferences != nil {
			s.OwnerReferences = make([]metav1.OwnerReference, len(ksvc.OwnerReferences))
			copy(s.OwnerReferences, ksvc.OwnerReferences)
		}

		// preserve status to avoid resetting conditions
		s.Status = ksvc.Status
	}
}
