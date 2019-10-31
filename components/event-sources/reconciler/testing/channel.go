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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/apis/duck/v1beta1"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/ptr"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
)

// NewChannel creates a Channel object.
func NewChannel(ns, name string, opts ...ChannelOption) *messagingv1alpha1.Channel {
	ch := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
	WithChannelController(name)(ch)
	for _, opt := range opts {
		opt(ch)
	}

	return ch
}

type ChannelOption func(*messagingv1alpha1.Channel)

func WithChannelController(srcName string) ChannelOption {
	return func(ch *messagingv1alpha1.Channel) {
		gvk := sourcesv1alpha1.HTTPSourceGVK()

		ch.OwnerReferences = []metav1.OwnerReference{{
			APIVersion:         gvk.GroupVersion().String(),
			Kind:               gvk.Kind,
			Name:               srcName,
			UID:                uid,
			Controller:         ptr.Bool(true),
			BlockOwnerDeletion: ptr.Bool(true),
		}}
	}
}

// WithSinkURI sets the Channel's sink URI.
func WithSinkURI(uri string) ChannelOption {
	return func(ch *messagingv1alpha1.Channel) {
		ch.Status.Address = &v1alpha1.Addressable{
			Addressable: v1beta1.Addressable{
				URL: &apis.URL{
					Host:uri,
				},
			},
		}
	}
}
