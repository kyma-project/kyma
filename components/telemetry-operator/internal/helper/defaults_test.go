package helper

import (
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestDefaultInputWithEmptySelectorsAndIncludeOptionFalse(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{}},
		},
	}

	result := SetDefaults(logPipeline)

	require.True(t, result)
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "kyma-system")
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "kube-system")
}

func TestDefaultInputWithEmptySelectorsAndIncludeOptionTrue(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{IncludeSystemNamespaces: true}},
		},
	}

	result := SetDefaults(logPipeline)

	require.False(t, result)
	require.Empty(t, logPipeline.Spec.Input.Application.Namespaces)
	require.Empty(t, logPipeline.Spec.Input.Application.ExcludeNamespaces)
}

func TestDefaultInputWithSetNamespacesAndIncludeOptionFalse(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces: []string{"namespace1", "namespace2"},
			}},
		},
	}

	result := SetDefaults(logPipeline)

	require.False(t, result)
	require.Contains(t, logPipeline.Spec.Input.Application.Namespaces, "namespace1")
	require.Contains(t, logPipeline.Spec.Input.Application.Namespaces, "namespace2")
	require.NotContains(t, logPipeline.Spec.Input.Application.Namespaces, "kyma-system")
	require.NotContains(t, logPipeline.Spec.Input.Application.Namespaces, "kube-system")
	require.Empty(t, logPipeline.Spec.Input.Application.ExcludeNamespaces)
}

func TestDefaultInputWithSetNamespacesAndIncludeOptionTrue(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces:              []string{"namespace1", "namespace2"},
				IncludeSystemNamespaces: true,
			}},
		},
	}

	result := SetDefaults(logPipeline)

	require.True(t, result)
	require.Contains(t, logPipeline.Spec.Input.Application.Namespaces, "namespace1")
	require.Contains(t, logPipeline.Spec.Input.Application.Namespaces, "namespace2")
	require.Contains(t, logPipeline.Spec.Input.Application.Namespaces, "kyma-system")
	require.Contains(t, logPipeline.Spec.Input.Application.Namespaces, "kube-system")
	require.Empty(t, logPipeline.Spec.Input.Application.ExcludeNamespaces)
}

func TestDefaultInputWithSetExcludedNamespacesAndIncludeOptionFalse(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				ExcludeNamespaces: []string{"namespace1", "namespace2"},
			}},
		},
	}

	result := SetDefaults(logPipeline)

	require.True(t, result)
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "namespace1")
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "namespace2")
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "kyma-system")
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "kube-system")
	require.Empty(t, logPipeline.Spec.Input.Application.Namespaces)
}

func TestDefaultInputWithSetExcludedNamespacesAndIncludeOptionTrue(t *testing.T) {
	logPipeline := &v1alpha1.LogPipeline{
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				ExcludeNamespaces:       []string{"namespace1", "namespace2"},
				IncludeSystemNamespaces: true,
			}},
		},
	}

	result := SetDefaults(logPipeline)

	require.False(t, result)
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "namespace1")
	require.Contains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "namespace2")
	require.NotContains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "kyma-system")
	require.NotContains(t, logPipeline.Spec.Input.Application.ExcludeNamespaces, "kube-system")
	require.Empty(t, logPipeline.Spec.Input.Application.Namespaces)
}
