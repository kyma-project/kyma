package fluentbit

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestBuildSection(t *testing.T) {
	expected := `[PARSER]
    Name   dummy_test
    Format   regex
    Regex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$

`
	content := "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
	actual := BuildConfigSection(ParserConfigHeader, content)

	require.Equal(t, expected, actual, "Fluent Bit Config Build is invalid")
}

func TestBuildSectionWithWrongIndentation(t *testing.T) {
	expected := `[PARSER]
    Name   dummy_test
    Format   regex
    Regex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$

`
	content := "Name   dummy_test   \n  Format   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
	actual := BuildConfigSection(ParserConfigHeader, content)

	require.Equal(t, expected, actual, "Fluent Bit config indentation has not been fixed")
}

func TestBuildConfigSectionFromMap(t *testing.T) {
	expected := `[FILTER]
    Key_A Value_A
    Key_B Value_B

`

	content := map[string]string{
		"Key_A": "Value_A",
		"Key_B": "Value_B",
	}
	actual := buildConfigSectionFromMap(FilterConfigHeader, content)

	require.Equal(t, expected, actual, "Fluent Bit config Build from Map is invalid")

}

func TestGeneratePermanentFilter(t *testing.T) {
	expected := `
name                  record_modifier
match                 foo.*
Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}`

	actual := generatePermanentFilter("foo")
	require.Equal(t, expected, actual, "Fluent Bit Permanent parser config is invalid")
}

func TestFilter(t *testing.T) {
	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $log "^.*$" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    match foo.*
    name grep
    regex log aa

[OUTPUT]
    allow_duplicated_headers true
    format json
    host localhost
    match foo.*
    name http
    port 443
    storage.total_limit_size 1G
    tls on
    tls.verify on
    uri /customindex/kyma

`
	filters := []telemetryv1alpha1.Filter{
		{
			Custom: `
	name grep
    regex log aa
`,
		},
	}
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: filters,
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	actual, err := MergeSectionsConfig(logPipeline, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestCustomOutput(t *testing.T) {
	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $log "^.*$" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[OUTPUT]
    match foo.*
    name http
    storage.total_limit_size 1G

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    name               http`,
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	actual, err := MergeSectionsConfig(logPipeline, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestHTTPOutput(t *testing.T) {
	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $log "^.*$" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[OUTPUT]
    allow_duplicated_headers true
    format json
    host localhost
    http_passwd password
    http_user user
    match foo.*
    name http
    port 443
    storage.total_limit_size 1G
    tls on
    tls.verify on
    uri /customindex/kyma

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
					User: telemetryv1alpha1.ValueType{
						Value: "user",
					},
					Password: telemetryv1alpha1.ValueType{
						Value: "password",
					},
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	actual, err := MergeSectionsConfig(logPipeline, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestHTTPOutputWithSecretReference(t *testing.T) {
	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $log "^.*$" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[OUTPUT]
    allow_duplicated_headers true
    format json
    host localhost
    http_passwd ${FOO_MY_NAMESPACE_SECRET_KEY}
    http_user user
    match foo.*
    name http
    port 443
    storage.total_limit_size 1G
    tls on
    tls.verify on
    uri /my-uri

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
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
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	actual, err := MergeSectionsConfig(logPipeline, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestResolveDirectValue(t *testing.T) {
	value := telemetryv1alpha1.ValueType{
		Value: "test",
	}
	resolved, err := resolveValue(value, "pipeline")
	require.NoError(t, err)
	require.Equal(t, resolved, value.Value)
}

func TestResolveSecretValue(t *testing.T) {
	value := telemetryv1alpha1.ValueType{
		ValueFrom: telemetryv1alpha1.ValueFromType{
			SecretKey: telemetryv1alpha1.SecretKeyRef{
				Name:      "test-name",
				Key:       "test-key",
				Namespace: "test-namespace",
			},
		},
	}
	resolved, err := resolveValue(value, "pipeline")
	require.NoError(t, err)
	require.Equal(t, resolved, "${PIPELINE_TEST_NAMESPACE_TEST_NAME_TEST_KEY}")
}
