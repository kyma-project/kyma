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
	"testing"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
)

func TestNewChannel(t *testing.T) {
	const (
		ns   = "testns"
		name = "test"
		img  = "registry/image:tag"
	)

	testHTTPSrc := &sourcesv1alpha1.HTTPSource{ObjectMeta: metav1.ObjectMeta{
		Namespace: ns,
		Name:      name,
		UID:       "00000000-0000-0000-0000-000000000000",
	}}
	testOwner := testHTTPSrc.ToOwner()

	t.Run("Without existing Service", func(t *testing.T) {
		ksvc := NewService(ns, name,
			WithContainerImage(img),
			WithServiceControllerRef(testOwner),
		)

		expectKsvc := &servingv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       ns,
				Name:            name,
				OwnerReferences: []metav1.OwnerReference{*testOwner},
			},
			Spec: servingv1.ServiceSpec{
				ConfigurationSpec: servingv1.ConfigurationSpec{
					Template: servingv1.RevisionTemplateSpec{
						Spec: servingv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: img,
								}},
							},
						},
					},
				},
			},
		}

		if d := cmp.Diff(expectKsvc, ksvc); d != "" {
			t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
		}
	})

	t.Run("Attributes from existing Service are preserved", func(t *testing.T) {
		existingKsvc := &servingv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       ns,
				Name:            name,
				ResourceVersion: "1",
				Annotations: map[string]string{
					knativeServingAnnotations[0]: "some-user",
					"another-annotation":         "some-value",
				},
			},
			Status: servingv1.ServiceStatus{
				Status: duckv1.Status{
					ObservedGeneration: 1,
				},
			},
		}

		ksvc := NewService(ns, name,
			WithExistingService(existingKsvc),
			WithServiceControllerRef(testOwner),
		)

		expectKsvc := &servingv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       ns,
				Name:            name,
				ResourceVersion: "1",
				Annotations: map[string]string{
					knativeServingAnnotations[0]: "some-user",
				},
				OwnerReferences: []metav1.OwnerReference{*testOwner},
			},
			Status: servingv1.ServiceStatus{
				Status: duckv1.Status{
					ObservedGeneration: 1,
				},
			},
		}

		if d := cmp.Diff(expectKsvc, ksvc); d != "" {
			t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
		}
	})

	t.Run("Owner references", func(t *testing.T) {
		testCases := map[string]struct {
			existing []metav1.OwnerReference
			expect   []metav1.OwnerReference
		}{
			"No existing": {
				existing: nil,
				expect: []metav1.OwnerReference{
					*testOwner,
				},
			},
			"Existing without controller": {
				existing: []metav1.OwnerReference{
					{Name: "dummy"},
				},
				expect: []metav1.OwnerReference{
					{Name: "dummy"},
					*testOwner,
				},
			},
			"Existing with controller": {
				existing: []metav1.OwnerReference{
					{Name: "dummy", Controller: ptr.Bool(true)},
				},
				expect: []metav1.OwnerReference{
					*testOwner,
				},
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				existingKsvc := &servingv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: tc.existing,
					},
				}

				ksvc := NewService(ns, name,
					WithExistingService(existingKsvc),
					WithServiceControllerRef(testOwner),
				)

				if d := cmp.Diff(tc.expect, ksvc.OwnerReferences); d != "" {
					t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
				}
			})
		}
	})
}
