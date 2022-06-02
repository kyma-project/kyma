package fluentbit

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateEmpty(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{},
	}

	err := pluginValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateAllowedFilters(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{"grep", "lua", "multiline"}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{
					Content: `
    Name    grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateAllowedUpperCaseFilters(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{"grep", "lua", "multiline"}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{
					Content: `
    Name    Grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateForbiddenFilters(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{"lua", "multiline"}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{
					Content: `
    Name    grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.Error(t, err)
}

func TestValidateAllowedOutputs(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{"loki", "http"})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    http
    Match   *`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateForbiddenOutputs(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{"loki", "http"})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    es
    Match   *`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.Error(t, err)
}

func TestValidateUnnamedOutputs(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{"loki", "http"})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Match   *`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.Error(t, err)
}

func TestValidateAllowAll(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    http
Match   *`,
				},
			},
			Filters: []telemetryv1alpha1.Filter{
				{
					Content: `
    Name    grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}

	err := pluginValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestEnableAllPlugins(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{"lua", "multiline"}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{
					Content: `
    Name    grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}
	logPipeline.Spec.EnableUnsupportedPlugins = true

	err := pluginValidator.Validate(logPipeline)
	require.NoError(t, err)
}
