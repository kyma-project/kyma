package serverless

import (
	"fmt"
	"math/rand"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestFunction(namespace, name string, minReplicas, maxReplicas int) *serverlessv1alpha1.Function {
	one := int32(minReplicas)
	two := int32(maxReplicas)
	suffix := rand.Int()

	return &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, suffix),
			Namespace: namespace,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Type:    serverlessv1alpha1.SourceTypeGit,
			Source:  fmt.Sprintf("%s-%d", name, suffix),
			Runtime: serverlessv1alpha1.Nodejs12,
			Repository: serverlessv1alpha1.Repository{
				BaseDir:   "/",
				Reference: "main",
			},
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
			Resources:   corev1.ResourceRequirements{},
			MinReplicas: &one,
			MaxReplicas: &two,
			Labels: map[string]string{
				testBindingLabel1: "foobar",
				testBindingLabel2: testBindingLabelValue,
				"foo":             "bar",
			},
		},
	}
}

func newTestRepository(name, namespace string, auth *serverlessv1alpha1.RepositoryAuth) *serverlessv1alpha1.GitRepository {
	return &serverlessv1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: serverlessv1alpha1.GitRepositorySpec{
			URL:  "https://mock.repo/kyma/test",
			Auth: auth,
		},
	}
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
