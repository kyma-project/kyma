package secret

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
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
