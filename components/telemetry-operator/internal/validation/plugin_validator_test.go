package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateEmpty(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{Spec: telemetryv1alpha1.LogPipelineSpec{}}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	require.Equal(t, "error validating output plugin: no output is defined, you must define one output", err.Error())
}

func TestValidateForbiddenFilters(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: ``,
			},
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    lua
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{"lua"}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin 'lua' is not supported. ")
}

func TestValidateFiltersContainMatchCondition(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin 'grep' contains match condition. Match conditions are forbidden")
}

func TestValidateForbiddenOutputs(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    loki`,
			},
		},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{"loki", "http"})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin 'loki' is not supported. ")
}

func TestValidateOutputContainsMatchCondition(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    es
    Match   *`,
			},
		},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin 'es' contains match condition. Match conditions are forbidden")
}

func TestValidateUnnamedOutputs(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
	Regex   $kubernetes['labels']['app'] my-deployment`,
			},
		},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{"loki", "http"})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration section does not have name attribute")
}

func TestValidateOutputsAndFiltersContainMatchCondition(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    http`,
			},
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    grep
    Match   *
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin 'grep' contains match condition. Match conditions are forbidden")
}

func TestValidatePipelineCreation(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    http`,
			},
		},
	}
	logPipeline.ObjectMeta = metav1.ObjectMeta{Name: "foo"}
	logPipelines := &telemetryv1alpha1.LogPipelineList{}

	sut := NewPluginValidator([]string{}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.NoError(t, err)
}

func TestDeniedFilterPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    lua
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}
	logPipeline.Name = "foo"
	logPipelines := &telemetryv1alpha1.LogPipelineList{}

	sut := NewPluginValidator([]string{"lua", "multiline"}, []string{})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error validating filter plugins: plugin 'lua' is not supported. ")
}

func TestDeniedOutputPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    lua
    Regex   $kubernetes['labels']['app'] my-deployment`,
			},
		},
	}
	logPipeline.Name = "foo"
	logPipelines := &telemetryv1alpha1.LogPipelineList{}

	sut := NewPluginValidator([]string{}, []string{"lua", "multiline"})
	err := sut.Validate(logPipeline, logPipelines)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error validating output plugin: plugin 'lua' is not supported. ")
}

func TestContainsNoOutputPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{},
		}}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{})
	result := sut.Validate(logPipeline, logPipelines)

	require.Error(t, result)
	require.Contains(t, result.Error(), "no output is defined, you must define one output")
}

func TestContainsMultipleOutputPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `Name	http`,
				HTTP: telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
				},
			},
		}}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{*logPipeline}}

	sut := NewPluginValidator([]string{}, []string{})
	result := sut.Validate(logPipeline, logPipelines)

	require.Error(t, result)
	require.Contains(t, result.Error(), "multiple output plugins are defined, you must define only one output")
}

func TestContainsCustomPluginWithCustomFilter(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    some-filter`,
				},
			},
		},
	}

	sut := NewPluginValidator([]string{}, []string{})
	result := sut.ContainsCustomPlugin(logPipeline)

	require.True(t, result)
}

func TestContainsCustomPluginWithCustomOutput(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    some-output`,
			},
		},
	}

	sut := NewPluginValidator([]string{}, []string{})
	result := sut.ContainsCustomPlugin(logPipeline)

	require.True(t, result)
}

func TestContainsCustomPluginWithoutAny(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{Spec: telemetryv1alpha1.LogPipelineSpec{}}

	sut := NewPluginValidator([]string{}, []string{})
	result := sut.ContainsCustomPlugin(logPipeline)

	require.False(t, result)
}

func TestValidHostname(t *testing.T) {
	require.True(t, validHostname("localhost"))
	require.True(t, validHostname("logging-loki"))
	require.True(t, validHostname("logging-loki.kyma-system.svc.cluster.local"))
	require.False(t, validHostname("https://logging-loki.kyma-system.svc.cluster.local"))
	require.False(t, validHostname("logging-loki.kyma-system.svc.cluster.local:443"))
	require.False(t, validHostname("!@#$$%"))
}

func TestValidLokiURL(t *testing.T) {
	require.True(t, validURL("http://logging-loki:3100/loki/api/v1/push"))
	require.False(t, validURL("http//abc.abc"))
}

func TestValidateHTTPOutput(t *testing.T) {
	output := telemetryv1alpha1.HTTPOutput{
		Host: telemetryv1alpha1.ValueType{
			Value: "localhost",
		},
	}
	require.NoError(t, validateHTTPOutput(output))

	output = telemetryv1alpha1.HTTPOutput{
		Host: telemetryv1alpha1.ValueType{
			Value: "localhost",
		},
		URI: "/my-path",
	}
	require.NoError(t, validateHTTPOutput(output))

	output = telemetryv1alpha1.HTTPOutput{
		Host: telemetryv1alpha1.ValueType{
			Value: "http://localhost",
		},
		URI: "/my-path",
	}
	err := validateHTTPOutput(output)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid hostname")

	output = telemetryv1alpha1.HTTPOutput{
		Host: telemetryv1alpha1.ValueType{
			Value: "localhost",
		},
		URI: "broken-uri",
	}
	err = validateHTTPOutput(output)
	require.Error(t, err)
	require.Contains(t, err.Error(), "uri has to start with /")

	output = telemetryv1alpha1.HTTPOutput{
		URI: "/my-path",
	}
	err = validateHTTPOutput(output)
	require.Error(t, err)
	require.Contains(t, err.Error(), "http output needs to have a host configured")
}
