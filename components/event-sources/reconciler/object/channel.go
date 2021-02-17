package object

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

// NewChannel creates a Channel object.
func NewChannel(ns, name string, opts ...ObjectOption) *messagingv1alpha1.Channel {
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
