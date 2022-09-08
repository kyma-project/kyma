package builder

import (
	"testing"

	"github.com/stretchr/testify/require"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func TestCreateOutputSectionWithCustomOutput(t *testing.T) {
	expected := `[OUTPUT]
    name                     null
    match                    foo.*
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
	pipelineConfig := PipelineDefaults{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateOutputSectionWithHTTPOutput(t *testing.T) {
	expected := `[OUTPUT]
    name                     http
    match                    foo.*
    alias                    foo - http
    allow_duplicated_headers true
    format                   yaml
    host                     localhost
    http_passwd              password
    http_user                user
    port                     1234
    storage.total_limit_size 1G
    tls                      on
    tls.verify               on
    uri                      /customindex/kyma

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: &telemetryv1alpha1.HTTPOutput{
					Dedot:    true,
					Port:     "1234",
					Host:     telemetryv1alpha1.ValueType{Value: "localhost"},
					User:     telemetryv1alpha1.ValueType{Value: "user"},
					Password: telemetryv1alpha1.ValueType{Value: "password"},
					URI:      "/customindex/kyma",
					Format:   "yaml",
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineDefaults{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateOutputSectionWithHTTPOutputWithSecretReference(t *testing.T) {
	expected := `[OUTPUT]
    name                     http
    match                    foo.*
    alias                    foo - http
    allow_duplicated_headers true
    format                   json
    host                     localhost
    http_passwd              ${FOO_MY_NAMESPACE_SECRET_KEY}
    http_user                user
    port                     443
    storage.total_limit_size 1G
    tls                      on
    tls.verify               on
    uri                      /my-uri

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: &telemetryv1alpha1.HTTPOutput{
					Dedot: true,
					URI:   "/my-uri",
					Host:  telemetryv1alpha1.ValueType{Value: "localhost"},
					User:  telemetryv1alpha1.ValueType{Value: "user"},
					Password: telemetryv1alpha1.ValueType{
						ValueFrom: &telemetryv1alpha1.ValueFromSource{
							SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
								Name:      "secret",
								Key:       "key",
								Namespace: "my-namespace",
							},
						},
					},
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineDefaults{FsBufferLimit: "1G"}

	actual := createOutputSection(logPipeline, pipelineConfig)
	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateOutputSectionWithLokiOutput(t *testing.T) {
	expected := `[OUTPUT]
    name                     grafana-loki
    match                    foo.*
    alias                    foo - grafana-loki
    labelmappath             /fluent-bit/etc/loki-labelmap.json
    labels                   {cluster-id="123", job="telemetry-fluent-bit"}
    lineformat               json
    loglevel                 warn
    removekeys               key1, key2
    storage.total_limit_size 1G
    url                      http:loki:3100

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Loki: &telemetryv1alpha1.LokiOutput{
					URL: telemetryv1alpha1.ValueType{Value: "http:loki:3100"},
					Labels: map[string]string{
						"job":        "telemetry-fluent-bit",
						"cluster-id": "123"},
					RemoveKeys: []string{"key1", "key2"},
				},
			},
		},
	}

	logPipeline.Name = "foo"
	pipelineConfig := PipelineDefaults{FsBufferLimit: "1G"}

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
		ValueFrom: &telemetryv1alpha1.ValueFromSource{
			SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
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
