package serverless

import (
	"fmt"
	"math/rand"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestGitFunction(namespace, name string, auth *serverlessv1alpha2.RepositoryAuth, minReplicas, maxReplicas int, continuousGitCheckout bool) *serverlessv1alpha2.Function {
	one := int32(minReplicas)
	two := int32(maxReplicas)
	//nolint:gosec
	suffix := rand.Int()
	annotations := map[string]string{}
	if continuousGitCheckout {
		annotations[continuousGitCheckoutAnnotation] = "true"
	}

	return &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%d", name, suffix),
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			Source: serverlessv1alpha2.Source{
				GitRepository: &serverlessv1alpha2.GitRepositorySource{
					URL: "https://mock.repo/kyma/test",
					Repository: serverlessv1alpha2.Repository{
						BaseDir:   "/",
						Reference: "main",
					},
					Auth: auth,
				},
			},
			Runtime: serverlessv1alpha2.NodeJs18,
			Env: []corev1.EnvVar{
				{
					Name:  "TEST_1",
					Value: "VAL_1",
				},
				{
					Name:  "TEST_2",
					Value: "VAL_2",
				},
			},
			ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
				Function: &serverlessv1alpha2.ResourceRequirements{
					Resources: &corev1.ResourceRequirements{},
				},
			},
			ScaleConfig: &serverlessv1alpha2.ScaleConfig{
				MinReplicas: &one,
				MaxReplicas: &two,
			},
			Labels: map[string]string{
				testBindingLabel1: "foobar",
				testBindingLabel2: testBindingLabelValue,
				"foo":             "bar",
			},
			Annotations: map[string]string{
				"foo": "bar",
			},
			SecretMounts: []serverlessv1alpha2.SecretMount{
				{
					SecretName: "secret-name-1",
					MountPath:  "/mount/path/1",
				},
				{
					SecretName: "secret-name-2",
					MountPath:  "/mount/path/2",
				},
			},
		},
	}
}

func newFixFunction(namespace, name string, minReplicas, maxReplicas int) *serverlessv1alpha2.Function {
	one := int32(minReplicas)
	two := int32(maxReplicas)
	//nolint:gosec
	suffix := rand.Int()

	return &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, suffix),
			Namespace: namespace,
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			Source: serverlessv1alpha2.Source{
				Inline: &serverlessv1alpha2.InlineSource{
					Source:       "module.exports = {main: function(event, context) {return 'Hello World.'}}",
					Dependencies: "   ",
				},
			},
			Runtime: serverlessv1alpha2.NodeJs18,
			Env: []corev1.EnvVar{
				{
					Name:  "TEST_1",
					Value: "VAL_1",
				},
				{
					Name:  "TEST_2",
					Value: "VAL_2",
				},
			},
			ScaleConfig: &serverlessv1alpha2.ScaleConfig{
				MinReplicas: &one,
				MaxReplicas: &two,
			},
			Labels: map[string]string{
				testBindingLabel1: "foobar",
				testBindingLabel2: testBindingLabelValue,
				"foo":             "bar",
			},
			Annotations: map[string]string{
				"foo": "bar",
			},
			SecretMounts: []serverlessv1alpha2.SecretMount{
				{
					SecretName: "secret-name-1",
					MountPath:  "/mount/path/1",
				},
				{
					SecretName: "secret-name-2",
					MountPath:  "/mount/path/2",
				},
			},
		},
	}
}

func newFixFunctionWithCustomImage(namespace, name, runtimeImageOverride string, minReplicas, maxReplicas int) *serverlessv1alpha2.Function {
	fn := newFixFunction(namespace, name, minReplicas, maxReplicas)
	fn.Spec.RuntimeImageOverride = runtimeImageOverride
	return fn
}

func newFixFunctionWithFunctionResourceProfile(namespace, name, profile string) *serverlessv1alpha2.Function {
	fn := newFixFunction(namespace, name, 1, 1)
	fn.Spec.ResourceConfiguration = &serverlessv1alpha2.ResourceConfiguration{
		Function: &serverlessv1alpha2.ResourceRequirements{Profile: profile},
	}
	return fn
}

func newFixFunctionWithBuildResourceProfile(namespace, name, profile string) *serverlessv1alpha2.Function {
	fn := newFixFunction(namespace, name, 1, 1)
	fn.Spec.ResourceConfiguration = &serverlessv1alpha2.ResourceConfiguration{
		Build: &serverlessv1alpha2.ResourceRequirements{Profile: profile},
	}
	return fn
}

func newFixFunctionWithRuntime(namespace, name string, runtime serverlessv1alpha2.Runtime) *serverlessv1alpha2.Function {
	fn := newFixFunction(namespace, name, 1, 1)
	fn.Spec.Runtime = runtime
	return fn
}

func newFixFunctionWithCustomFunctionResource(namespace, name string, resources *corev1.ResourceRequirements) *serverlessv1alpha2.Function {
	fn := newFixFunction(namespace, name, 1, 1)
	fn.Spec.ResourceConfiguration = &serverlessv1alpha2.ResourceConfiguration{
		Function: &serverlessv1alpha2.ResourceRequirements{Resources: resources},
	}
	return fn
}

func newFixFunctionWithCustomBuildResource(namespace, name string, resources *corev1.ResourceRequirements) *serverlessv1alpha2.Function {
	fn := newFixFunction(namespace, name, 1, 1)
	fn.Spec.ResourceConfiguration = &serverlessv1alpha2.ResourceConfiguration{
		Build: &serverlessv1alpha2.ResourceRequirements{Resources: resources},
	}
	return fn
}

func newTestSecret(name, namespace string, stringData map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: stringData,
	}
}
