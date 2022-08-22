package builder

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestCreateOutputSectionWithCustomOutput(t *testing.T) {
	expected := `[OUTPUT]
    match                    foo.*
    name                     null
    storage.total_limit_size 1G

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    name null`,
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateOutputSectionWithHTTPOutput(t *testing.T) {
	expected := `[OUTPUT]
    allow_duplicated_headers true
    format                   json
    host                     localhost
    http_passwd              password
    http_user                user
    match                    foo.*
    name                     http
    port                     443
    storage.total_limit_size 1G
    tls                      on
    tls.verify               on
    uri                      /customindex/kyma

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: true,
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
					User: telemetryv1alpha1.ValueType{
						Value: "user",
					},
					Password: telemetryv1alpha1.ValueType{
						Value: "password",
					},
					URI: "/customindex/kyma",
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateOutputSectionWithHTTPOutputWithSecretReference(t *testing.T) {
	expected := `[OUTPUT]
    allow_duplicated_headers true
    format                   json
    host                     localhost
    http_passwd              ${FOO_MY_NAMESPACE_SECRET_KEY}
    http_user                user
    match                    foo.*
    name                     http
    port                     443
    storage.total_limit_size 1G
    tls                      on
    tls.verify               on
    uri                      /my-uri

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: true,
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
					User: telemetryv1alpha1.ValueType{
						Value: "user",
					},
					Password: telemetryv1alpha1.ValueType{
						ValueFrom: telemetryv1alpha1.ValueFromType{
							SecretKey: telemetryv1alpha1.SecretKeyRef{
								Name:      "secret",
								Key:       "key",
								Namespace: "my-namespace",
							},
						},
					},
					URI: "/my-uri",
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateOutputSectionWithLokiOutput(t *testing.T) {
	expected := `[OUTPUT]
    alias                    foo
    labelMapPath             /fluent-bit/etc/loki-labelmap.json
    labels                   {cluster-id="123", job="telemetry-fluent-bit"}
    lineformat               json
    loglevel                 warn
    match                    foo.*
    name                     grafana-loki
    removeKeys               key1, key2
    storage.total_limit_size 1G
    url                      http:loki:3100

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Loki: telemetryv1alpha1.LokiOutput{
					URL: telemetryv1alpha1.ValueType{
						Value: "http:loki:3100",
					},
					Labels: map[string]string{
						"job":        "telemetry-fluent-bit",
						"cluster-id": "123"},
					RemoveKeys: []string{"key1", "key2"},
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestResolveValueWithValue(t *testing.T) {
	value := telemetryv1alpha1.ValueType{
		Value: "test",
	}
	resolved := resolveValue(value, "pipeline")
	require.NotEmpty(t, resolved)
	require.Equal(t, resolved, value.Value)
}

func TestResolveValueWithSecretKeyRef(t *testing.T) {
	value := telemetryv1alpha1.ValueType{
		ValueFrom: telemetryv1alpha1.ValueFromType{
			SecretKey: telemetryv1alpha1.SecretKeyRef{
				Name:      "test-name",
				Key:       "test-key",
				Namespace: "test-namespace",
			},
		},
	}
	resolved := resolveValue(value, "pipeline")
	require.NotEmpty(t, resolved)
	require.Equal(t, resolved, "${PIPELINE_TEST_NAMESPACE_TEST_NAME_TEST_KEY}")
}
