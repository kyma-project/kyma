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

// Package taskrun contains utilities for handling TaskRuns (Tekton Pipelines).
package taskrun

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/test/e2e/framework"

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

// Copied from github.com/tektoncd/pipeline/pkg/status to avoid an import conflict:
// -> github.com/tektoncd/pipeline/pkg/reconciler/taskrun/resources
//  -> github.com/tektoncd/pipeline/pkg/reconciler/taskrun/entrypoint
//   -> github.com/google/go-containerregistry/pkg/authn/k8schain
//    -> k8s.io/kubernetes/pkg/credentialprovider (v1.11.10)
const (
	// ReasonSucceeded indicates that the reason for the finished status is that all of the steps
	// completed successfully
	ReasonSucceeded = "Succeeded"

	// ReasonFailed indicates that the reason for the failure status is unknown or that one of the steps failed
	ReasonFailed = "Failed"
)

// FromUnstructured converts an instance of Unstructured to a TaskRun object.
func FromUnstructured(tr *unstructured.Unstructured) *tektonv1alpha1.TaskRun {
	trObj := &tektonv1alpha1.TaskRun{}

	convertCtx := runtime.NewMultiGroupVersioner(tektonv1alpha1.SchemeGroupVersion)
	if err := scheme.Scheme.Convert(tr, trObj, convertCtx); err != nil {
		framework.Failf("Error converting Unstructured to TaskRun: %v", err)
	}

	return trObj
}
