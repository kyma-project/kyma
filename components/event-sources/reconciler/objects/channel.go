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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/eventing/pkg/apis/messaging"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
)

// List of annotations set on Knative Messaging objects by the Knative Eventing
// admission webhook.
var knativeMessagingAnnotations = []string{
	messaging.GroupName + apis.CreatorAnnotationSuffix,
	messaging.GroupName + apis.UpdaterAnnotationSuffix,
}

// NewChannel creates a Channel object.
func NewChannel(ns, name string, opts ...ChannelOption) *messagingv1alpha1.Channel {
	s := &messagingv1alpha1.Channel{
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

// ChannelOption is a functional option for Channel objects.
type ChannelOption func(*messagingv1alpha1.Channel)

// WithChannelControllerRef sets the controller reference of a Channel.
func WithChannelControllerRef(or *metav1.OwnerReference) ChannelOption {
	return func(s *messagingv1alpha1.Channel) {
		s.OwnerReferences = []metav1.OwnerReference{*or}
	}
}

// WithChannelLabel sets a label on a Channel object.
func WithChannelLabel(key, val string) ChannelOption {
	return func(s *messagingv1alpha1.Channel) {
		if s.Labels == nil {
			s.Labels = make(map[string]string, 1)
		}
		s.Labels[key] = val
	}
}

// ApplyExistingChannelAttributes copies some important attributes from a given
// source Channel to a destination Channel.
func ApplyExistingChannelAttributes(src, dst *messagingv1alpha1.Channel) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion

	// immutable Knative annotations must be preserved
	for _, ann := range knativeMessagingAnnotations {
		if val, ok := src.Annotations[ann]; ok {
			metav1.SetMetaDataAnnotation(&dst.ObjectMeta, ann, val)
		}
	}

	// preserve status to avoid resetting conditions
	dst.Status = src.Status
}
