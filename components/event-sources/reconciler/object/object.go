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

// Package object contains utilities for creating and comparing API objects.
package object

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ObjectOption is a functional option for API objects builders.
type ObjectOption func(metav1.Object)

// WithControllerRef sets the controller reference of an API object.
func WithControllerRef(or *metav1.OwnerReference) ObjectOption {
	return func(o metav1.Object) {
		o.SetOwnerReferences([]metav1.OwnerReference{*or})
	}
}

// WithLabel sets the value of an API object's label.
func WithLabel(key, val string) ObjectOption {
	return func(o metav1.Object) {
		lbls := o.GetLabels()
		if lbls == nil {
			lbls = make(map[string]string, 1)
			o.SetLabels(lbls)
		}
		lbls[key] = val
	}
}
