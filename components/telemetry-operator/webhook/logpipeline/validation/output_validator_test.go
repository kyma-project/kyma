package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestContainsNoOutputPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{},
		}}

	sut := NewOutputValidator()
	result := sut.Validate(logPipeline)

	require.Error(t, result)
	require.Contains(t, result.Error(), "no output is defined, you must define one output")
}

func TestContainsMultipleOutputPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `Name	http`,
				HTTP: &telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
				},
			},
		}}

	sut := NewOutputValidator()
	result := sut.Validate(logPipeline)

	require.Error(t, result)
	require.Contains(t, result.Error(), "multiple output plugins are defined, you must define only one output")
}

func TestDeniedOutputPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: v1.ObjectMeta{Name: "foo"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    lua`,
			},
		},
	}

	sut := NewOutputValidator("lua", "multiline")
	err := sut.Validate(logPipeline)

	require.Error(t, err)
	require.Contains(t, err.Error(), "plugin 'lua' is forbidden. ")
}

func TestValidateCustomOutput(t *testing.T) {
	outputValidator := NewOutputValidator()

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    name    http`,
			},
		},
	}

	err := outputValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateCustomHasForbiddenParameter(t *testing.T) {
	outputValidator := NewOutputValidator()

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    name    http
	storage.total_limit_size 10G`,
			},
		},
	}

	err := outputValidator.Validate(logPipeline)
	require.Error(t, err)
}

func TestValidateCustomOutputsContainsNoName(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
	Regex   .*`,
			},
		},
	}

	sut := NewOutputValidator()
	err := sut.Validate(logPipeline)

	require.Error(t, err)
	require.Contains(t, err.Error(), "configuration section does not have name attribute")
}

func TestValidateCustomOutputsContainsMatch(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    http
	Match   *
`,
			},
		},
	}

	sut := NewOutputValidator()
	err := sut.Validate(logPipeline)

	require.Error(t, err)
	require.Contains(t, err.Error(), "plugin 'http' contains match condition. Match conditions are forbidden")
}

func TestValidHostname(t *testing.T) {
	require.True(t, validHostname("localhost"))
	require.True(t, validHostname("logging-loki"))
	require.True(t, validHostname("logging-loki.kyma-system.svc.cluster.local"))
	require.False(t, validHostname("https://logging-loki.kyma-system.svc.cluster.local"))
	require.False(t, validHostname("logging-loki.kyma-system.svc.cluster.local:443"))
	require.False(t, validHostname("!@#$$%"))
}

func TestValidateHTTPOutput(t *testing.T) {
	tests := []struct {
		name      string
		given     *telemetryv1alpha1.HTTPOutput
		expectErr bool
	}{
		{
			name: "valid host",
			given: &telemetryv1alpha1.HTTPOutput{
				Host: telemetryv1alpha1.ValueType{
					Value: "localhost",
				},
			},
		},
		{
			name: "valid host with uri",
			given: &telemetryv1alpha1.HTTPOutput{
				Host: telemetryv1alpha1.ValueType{
					Value: "localhost",
				},
				URI: "/my-path",
			},
		},
		{
			name: "invalid host with schema",
			given: &telemetryv1alpha1.HTTPOutput{
				Host: telemetryv1alpha1.ValueType{
					Value: "http://localhost",
				},
				URI: "/my-path",
			},
			expectErr: true,
		},
		{
			name: "invalid uri",
			given: &telemetryv1alpha1.HTTPOutput{
				Host: telemetryv1alpha1.ValueType{
					Value: "localhost",
				},
				URI: "my-path",
			},
			expectErr: true,
		},
		{
			name: "no host",
			given: &telemetryv1alpha1.HTTPOutput{
				URI: "my-path",
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sut := NewOutputValidator()
			err := sut.Validate(&telemetryv1alpha1.LogPipeline{
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Output: telemetryv1alpha1.Output{
						HTTP: test.given,
					},
				},
			})

			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidLokiOutput(t *testing.T) {
	tests := []struct {
		name      string
		given     *telemetryv1alpha1.LokiOutput
		expectErr bool
	}{
		{
			name: "valid url",
			given: &telemetryv1alpha1.LokiOutput{
				URL: telemetryv1alpha1.ValueType{
					Value: "http://logging-loki:3100/loki/api/v1/push",
				},
			},
		},
		{
			name: "invalid url",
			given: &telemetryv1alpha1.LokiOutput{
				URL: telemetryv1alpha1.ValueType{
					Value: "http//abc.abc",
				},
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sut := NewOutputValidator()
			err := sut.Validate(&telemetryv1alpha1.LogPipeline{
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Output: telemetryv1alpha1.Output{
						Loki: test.given,
					},
				},
			})

			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
