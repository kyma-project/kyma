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
	pkgerrors "github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
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

	for _, opt := range opts {
		opt(ch)
	}

	return ch
}

// ChannelOption is a functional option for Channel objects.
type ChannelOption func(*messagingv1alpha1.Channel)

// WithChannelController sets the controller of a Channel.
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

// WithChannelSinkURI sets the sink URI of a Channel.
func WithChannelSinkURI(uri string) ChannelOption {
	return func(ch *messagingv1alpha1.Channel) {
		parsedURI, err := apis.ParseURL(uri)
		if err != nil {
			panic(pkgerrors.Wrap(err, "parsing Channel URL"))
		}

		ch.Status.Address = &duckv1alpha1.Addressable{
			Addressable: duckv1beta1.Addressable{
				URL: parsedURI,
			},
		}
	}
}
