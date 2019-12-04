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
	"testing"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func TestNewService(t *testing.T) {
	const img = "registry/image:tag"

	ksvc := NewService(tNs, tName,
		WithPort(8080),
		WithImage(img),
		WithEnvVar("TEST_ENV1", "val1"),
		WithPort(8081),
		WithProbe("/are/you/alive"),
		WithEnvVar("TEST_ENV2", "val2"),
	)

	expectKsvc := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
		Spec: servingv1alpha1.ServiceSpec{
			ConfigurationSpec: servingv1alpha1.ConfigurationSpec{
				Template: &servingv1alpha1.RevisionTemplateSpec{
					Spec: servingv1alpha1.RevisionSpec{
						RevisionSpec: servingv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: img,
									Ports: []corev1.ContainerPort{{
										ContainerPort: 8080,
									}, {
										ContainerPort: 8081,
									}},
									Env: []corev1.EnvVar{{
										Name:  "TEST_ENV1",
										Value: "val1",
									}, {
										Name:  "TEST_ENV2",
										Value: "val2",
									}},
									ReadinessProbe: &corev1.Probe{
										Handler: corev1.Handler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/are/you/alive",
											},
										},
									},
								}},
							},
						},
					},
				},
			},
		},
	}

	if d := cmp.Diff(expectKsvc, ksvc); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}

func TestAppApplyExistingServiceAttributes(t *testing.T) {
	existingKsvc := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "1",
			Annotations: map[string]string{
				knativeServingAnnotations[0]: "some-user",
				"another-annotation":         "some-value",
			},
		},
		Status: servingv1alpha1.ServiceStatus{
			Status: duckv1.Status{
				ObservedGeneration: 1,
			},
		},
	}

	// Service with empty spec, status, annotations, ...
	ksvc := NewService(tNs, tName)
	ApplyExistingServiceAttributes(existingKsvc, ksvc)

	expectKsvc := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "1",
			Annotations: map[string]string{
				knativeServingAnnotations[0]: "some-user",
			},
		},
		Status: servingv1alpha1.ServiceStatus{
			Status: duckv1.Status{
				ObservedGeneration: 1,
			},
		},
	}

	if d := cmp.Diff(expectKsvc, ksvc); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
