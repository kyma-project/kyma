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
)

// NewConfigMap creates a ConfigMap object.
func NewConfigMap(ns, name string, opts ...ConfigMapOption) *corev1.ConfigMap {
	cmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}

	for _, opt := range opts {
		opt(cmap)
	}

	return cmap
}

// ConfigMapOption is a functional option for ConfigMap objects.
type ConfigMapOption func(*corev1.ConfigMap)

// WithData sets the data of a ConfigMap.
func WithData(data map[string]string) ConfigMapOption {
	return func(cmap *corev1.ConfigMap) {
		cmap.Data = data
	}
}
