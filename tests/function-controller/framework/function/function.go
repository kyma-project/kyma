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

// Package function contains utilities for Functions tests.
package function

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/test/e2e/framework"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const functionManifestsPath = "framework/function/manifests"

const helloWorldJS = `module.exports = {
  main: function(event, context) {
    return 'Hello World'
  }
}`

// New returns a Function object initialized with valid specs.
func New(namespace, namePrefix string, options ...FunctionOption) *serverlessv1alpha1.Function {
	return &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: namePrefix,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Function: helloWorldJS,
			Deps:     "",
			Env:      []corev1.EnvVar{},
		},
		Status: serverlessv1alpha1.FunctionStatus{},
	}
}

// FunctionOption is a functional option type for serverlessv1alpha1.Function.
type FunctionOption func(fn *serverlessv1alpha1.Function)

// WithDefaults applies sane defaults to a Function.
func WithDefaults() FunctionOption {
	var opt FunctionOption = func(fn *serverlessv1alpha1.Function) {
		fn.Spec.FunctionContentType = "plaintext"
		fn.Spec.Size = "S"
		fn.Spec.Runtime = "nodejs8"
		fn.Spec.Timeout = 3
	}

	return opt
}

// CreateDockerfiles creates the ConfigMaps containing the Dockerfiles used to
// build Functions images.
func CreateDockerfiles(f *framework.Framework) {
	dockerfilesManifestPath := filepath.Join(functionManifestsPath, "dockerfiles.yaml")
	if _, err := f.CreateFromManifests(nil, dockerfilesManifestPath); err != nil {
		framework.Failf("Error creating registry objects from manifests: %v", err)
	}
}
