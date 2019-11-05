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

package function

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/test/e2e/framework"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

// ToUnstructured converts a Function object to its Unstructured representation.
func ToUnstructured(fn *serverlessv1alpha1.Function) *unstructured.Unstructured {
	fnUnstr := &unstructured.Unstructured{}

	convertCtx := runtime.NewMultiGroupVersioner(serverlessv1alpha1.SchemeGroupVersion)
	if err := scheme.Scheme.Convert(fn, fnUnstr, convertCtx); err != nil {
		framework.Failf("Error converting Function to Unstructured: %v", err)
	}

	return fnUnstr
}

// FromUnstructured converts an instance of Unstructured to a Function object.
func FromUnstructured(fn *unstructured.Unstructured) *serverlessv1alpha1.Function {
	fnObj := &serverlessv1alpha1.Function{}

	convertCtx := runtime.NewMultiGroupVersioner(serverlessv1alpha1.SchemeGroupVersion)
	if err := scheme.Scheme.Convert(fn, fnObj, convertCtx); err != nil {
		framework.Failf("Error converting Unstructured to Function: %v", err)
	}

	return fnObj
}
