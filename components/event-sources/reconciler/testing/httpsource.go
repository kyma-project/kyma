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
	"k8s.io/apimachinery/pkg/types"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
)

// used in conversion to OwnerReference
const uid = types.UID("00000000-0000-0000-0000-000000000000")

// default Spec fields
const (
	DefaultHTTPSource = "varkes"
)

func NewHTTPSource(ns, name string, opts ...HTTPSourceOption) *sourcesv1alpha1.HTTPSource {
	src := &sourcesv1alpha1.HTTPSource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			UID:       uid,
		},
	}

	for _, opt := range opts {
		opt(src)
	}

	if src.Spec.Source == "" {
		src.Spec.Source = DefaultHTTPSource
	}

	return src
}

type HTTPSourceOption func(s *sourcesv1alpha1.HTTPSource)

// WithInitConditions initializes all conditions to Unknown.
func WithInitConditions(s *sourcesv1alpha1.HTTPSource) {
	s.Status.InitializeConditions()
}

// WithDeployed sets the Deployed condition to True.
func WithDeployed(s *sourcesv1alpha1.HTTPSource) {
	s.Status.PropagateServiceReady(NewService("", "", WithServiceReady))
}

// WithNotDeployed sets the Deployed condition to False.
func WithNotDeployed(s *sourcesv1alpha1.HTTPSource) {
	s.Status.PropagateServiceReady(NewService("", ""))
}

// WithSink sets the SinkProvided condition to True.
func WithSink(uri string) HTTPSourceOption {
	return func(s *sourcesv1alpha1.HTTPSource) {
		s.Status.MarkSink(uri)
	}
}

// WithNoSink sets the SinkProvided condition to False.
func WithNoSink(s *sourcesv1alpha1.HTTPSource) {
	s.Status.MarkNoSink()
}
