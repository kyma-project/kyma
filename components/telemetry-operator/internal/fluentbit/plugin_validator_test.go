package fluentbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateEmpty(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{},
	}

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
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
    Match   tele.*
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
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
    Match   tele.*
    Regex   $kubernetes['labels']['app'] my-deployment`,
				},
			},
		},
	}

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
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

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
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
    Match   foo.*`,
				},
			},
		},
	}
	logPipeline.ObjectMeta = metav1.ObjectMeta{Name: "foo"}

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
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

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
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

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
	require.Error(t, err)
}

func TestValidateDisallowAll(t *testing.T) {
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

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	err := pluginValidator.Validate(logPipeline, logPipelines)
	assert.Contains(t, err.Error(), "filter plugin 'grep' with match condition '*' (match all) is not allowed")
}

func TestValidateMatchCondWithFirstLogPipeline(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{

		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    http
    Match   abc`,
				},
			},
		},
	}

	logPipeline.ObjectMeta = metav1.ObjectMeta{Name: "foo"}

	logPipelines := &telemetryv1alpha1.LogPipelineList{}

	err := pluginValidator.Validate(logPipeline, logPipelines)
	assert.Contains(t, err.Error(), "output plugin 'http' with match condition 'abc' is not allowed. Valid match conditions are: 'foo' (current logpipeline name)")
}

func TestValidateMatchCondWithExistingLogPipeline(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{})

	logPipeline1 := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    http
    Match   foo.*`,
				},
			},
		},
	}
	logPipeline1.ObjectMeta = metav1.ObjectMeta{Name: "foo"}

	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline1},
	}

	logPipeline2 := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    http
    Match   bar`,
				},
			},
		},
	}
	logPipeline2.ObjectMeta = metav1.ObjectMeta{Name: "bar"}

	err := pluginValidator.Validate(logPipeline2, logPipelines)
	assert.Contains(t, err.Error(), "output plugin 'http' with match condition 'bar' is not allowed. Valid match conditions are: 'bar' (current logpipeline name) or '[foo]' (other existing logpipelines names)")
}

func TestValidatePipelineCreation(t *testing.T) {
	pluginValidator := NewPluginValidator([]string{}, []string{})

	logPipeline := &telemetryv1alpha1.LogPipeline{

		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    Name    http
    Match   foo.*`,
				},
			},
		},
	}
	logPipeline.ObjectMeta = metav1.ObjectMeta{Name: "foo"}

	logPipelines := &telemetryv1alpha1.LogPipelineList{}

	err := pluginValidator.Validate(logPipeline, logPipelines)
	require.NoError(t, err)
}
