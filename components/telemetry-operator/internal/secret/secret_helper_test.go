package secret

import (
	"context"
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGenerateVariableName(t *testing.T) {
	expected := "PIPELINE_TEST_NAMESPACE_TEST_NAME_TEST_KEY_123"
	secretRef := telemetryv1alpha1.SecretKeyRef{
		Name:      "test-name",
		Key:       "TEST_KEY_123",
		Namespace: "test-namespace",
	}
	actual := GenerateVariableName(secretRef, "pipeline")
	require.Equal(t, expected, actual)
}

func TestGenerateVariableNameFromLowercase(t *testing.T) {
	expected := "PIPELINE_TEST_NAMESPACE_TEST_NAME_TEST_KEY_123"
	secretRef := telemetryv1alpha1.SecretKeyRef{
		Name:      "test-name",
		Key:       "test-key.123",
		Namespace: "test-namespace",
	}
	actual := GenerateVariableName(secretRef, "pipeline")
	require.Equal(t, expected, actual)
}

func TestValidateSecretExists(t *testing.T) {
	s := scheme.Scheme
	err := telemetryv1alpha1.AddToScheme(s)
	require.NoError(t, err)

	secretData := map[string][]byte{
		"host": []byte("my-host"),
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "referenced-secret",
			Namespace: "default",
		},
		Data: secretData,
	}

	mockClient := fake.NewClientBuilder().WithScheme(s).Build()
	helper := NewSecretHelper(mockClient)

	secretKeyRef := telemetryv1alpha1.SecretKeyRef{
		Name:      "referenced-secret",
		Key:       "host",
		Namespace: "default",
	}
	lp := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pipeline"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{
						ValueFrom: telemetryv1alpha1.ValueFromType{
							SecretKey: secretKeyRef,
						},
					},
				},
			},
		},
	}

	existing := helper.ValidateSecretsExist(context.Background(), &lp)
	require.False(t, existing)

	err = mockClient.Create(context.Background(), &secret)
	require.NoError(t, err)

	existing = helper.ValidateSecretsExist(context.Background(), &lp)
	require.True(t, existing)
}
