package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateWithValidInputNamespace(t *testing.T) {
	inputValidator := NewInputValidator()

	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: []string{"namespace-1", "namespace-2"},
			Containers: []string{"container-1"},
		},
	}

	err := inputValidator.Validate(&input)
	require.NoError(t, err)
}

func TestValidateWithValidInputExcludeNamespace(t *testing.T) {
	inputValidator := NewInputValidator()

	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			ExcludeNamespaces: []string{"namespace-1", "namespace-2"},
			ExcludeContainers: []string{"container-1"},
		},
	}

	err := inputValidator.Validate(&input)
	require.NoError(t, err)
}

func TestValidateWithInvalidNamespaceSelectors(t *testing.T) {
	inputValidator := NewInputValidator()

	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces:        []string{"namespace-1", "namespace-2"},
			ExcludeNamespaces: []string{"namespace-3"},
		},
	}

	err := inputValidator.Validate(&input)
	require.Error(t, err)
}

func TestValidateWithInvalidContainerSelectors(t *testing.T) {
	inputValidator := NewInputValidator()

	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Containers:        []string{"container-1", "container-2"},
			ExcludeContainers: []string{"container-3"},
		},
	}

	err := inputValidator.Validate(&input)
	require.Error(t, err)
}
