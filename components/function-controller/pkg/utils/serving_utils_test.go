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

package utils_test

import (
	"github.com/ghodss/yaml"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/components/function-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetServiceSpec(t *testing.T) {
	imageName := "foo-image"
	fn := serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Function:            "main() {}",
			FunctionContentType: "plaintext",
			Size:                "L",
			Runtime:             "nodejs8",
		},
	}

	rnInfo := &utils.RuntimeInfo{
		RegistryInfo: "test",
		AvailableRuntimes: []utils.RuntimesSupported{
			{
				ID:             "nodejs8",
				DockerfileName: "testnodejs8",
			},
		},
	}
	serviceSpec := utils.GetServiceSpec(imageName, fn, rnInfo)

	// Testing ConfigurationSpec
	if serviceSpec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image != "foo-image" {
		t.Fatalf("Expected image for RevisionTemplate.Spec.Container.Image: %v Got: %v", "foo-image", serviceSpec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image)
	}
	expectedEnv := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER",
			Value: "main",
		},
		{
			Name:  "MOD_NAME",
			Value: "handler",
		},
		{
			Name:  "FUNC_TIMEOUT",
			Value: "180",
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: "nodejs8",
		},
		{
			Name:  "FUNC_MEMORY_LIMIT",
			Value: "128Mi",
		},
		{
			Name:  "FUNC_PORT",
			Value: "8080",
		},
		{
			Name:  "NODE_PATH",
			Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
		},
	}
	if !compareEnv(t, expectedEnv, serviceSpec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Env) {
		expectedEnvStr, err := getString(expectedEnv)
		gotEnvStr, err := getString(expectedEnv)
		t.Fatalf("Expected value in Env: %v Got: %v", expectedEnvStr, gotEnvStr)
		if err != nil {
			t.Fatalf("Error while unmarshaling expectedBuildSpec: %v", err)
		}
	}
}

func compareEnv(t *testing.T, source, dest []corev1.EnvVar) bool {
	for i, _ := range source {
		found := false
		for j, _ := range dest {
			if source[i].Name == dest[j].Name && source[i].Value == dest[j].Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func getString(obj interface{}) (string, error) {
	output, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
