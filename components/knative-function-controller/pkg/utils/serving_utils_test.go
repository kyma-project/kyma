package utils_test

import (
	"github.com/ghodss/yaml"
	runtimev1alpha1 "github.com/kyma-project/kyma/components/knative-function-controller/pkg/apis/runtime/v1alpha1"
	"github.com/kyma-project/kyma/components/knative-function-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetServiceSpec(t *testing.T) {
	imageName := "foo-image"
	fn := runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: runtimev1alpha1.FunctionSpec{
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
				DockerFileName: "testnodejs8",
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
