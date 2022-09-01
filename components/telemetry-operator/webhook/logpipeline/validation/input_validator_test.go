package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateWithValidInputIncludes(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				Include: []string{"namespace-1", "namespace-2"},
			},
			Containers: telemetryv1alpha1.InputContainers{
				Include: []string{"container-1"},
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.NoError(t, err)
}

func TestValidateWithValidInputExcludes(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				Exclude: []string{"namespace-1", "namespace-2"},
			},
			Containers: telemetryv1alpha1.InputContainers{
				Exclude: []string{"container-1"},
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.NoError(t, err)
}

func TestValidateWithValidInputIncludeContainersSystemFlag(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				System: true,
			},
			Containers: telemetryv1alpha1.InputContainers{
				Include: []string{"container-1"},
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.NoError(t, err)
}

func TestValidateWithValidInputExcludeContainersSystemFlag(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				System: true,
			},
			Containers: telemetryv1alpha1.InputContainers{
				Exclude: []string{"container-1"},
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.NoError(t, err)
}

func TestValidateWithInvalidNamespaceSelectors(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				Include: []string{"namespace-1", "namespace-2"},
				Exclude: []string{"namespace-3"},
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.Error(t, err)
}

func TestValidateWithInvalidIncludeSystemFlag(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				Include: []string{"namespace-1", "namespace-2"},
				System:  true,
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.Error(t, err)
}

func TestValidateWithInvalidExcludeSystemFlag(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Namespaces: telemetryv1alpha1.InputNamespaces{
				Exclude: []string{"namespace-3"},
				System:  true,
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.Error(t, err)
}

func TestValidateWithInvalidContainerSelectors(t *testing.T) {
	input := telemetryv1alpha1.Input{
		Application: telemetryv1alpha1.ApplicationInput{
			Containers: telemetryv1alpha1.InputContainers{
				Include: []string{"container-1", "container-2"},
				Exclude: []string{"container-3"},
			},
		},
	}

	err := NewInputValidator().Validate(&input)
	require.Error(t, err)
}
