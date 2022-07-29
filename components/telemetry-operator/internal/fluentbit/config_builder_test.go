package fluentbit

import (
	"testing"

	fsmocks "github.com/kyma-project/kyma/components/telemetry-operator/internal/fs/mocks"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	actual := BuildConfigSectionFromMap(FilterConfigHeader, content)

	require.Equal(t, expected, actual, "Fluent Bit config Build from Map is invalid")

}

func TestGenerateEmitter(t *testing.T) {
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	expected := `
name                  rewrite_tag
match                 kube.*
Rule                  $log "^.*$" test.$TAG true
Emitter_Name          test
Emitter_Storage.type  filesystem
Emitter_Mem_Buf_Limit 10M`

	actual := generateEmitter(pipelineConfig, "test")
	require.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
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
    name                  rewrite_tag
    match                 kube.*
    Rule                  $log "^.*$" foo.$TAG true
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M

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
    name                  rewrite_tag
    match                 kube.*
    Rule                  $log "^.*$" foo.$TAG true
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M

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
    name                  rewrite_tag
    match                 kube.*
    Rule                  $log "^.*$" foo.$TAG true
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M

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
    name                  rewrite_tag
    match                 kube.*
    Rule                  $log "^.*$" foo.$TAG true
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M

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

func TestLokiOutputPlugin(t *testing.T) {
	lokiLabels := make(map[string]string)
	lokiLabels["job"] = "telemetry-fluent-bit"

	loki := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lokiFoo",
			Namespace: "lokiNs",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Loki: telemetryv1alpha1.LokiOutput{
					URL: telemetryv1alpha1.ValueType{
						Value: "http:loki:3100",
					},
					Labels:     lokiLabels,
					RemoveKeys: []string{"key1", "key2"},
				},
			},
		},
	}
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}
	fsWrapperMock := &fsmocks.Wrapper{}
	fsWrapperMock.On("CreateAndWrite", mock.Anything).Return(nil)
	res, err := generateLokiOutPut(loki, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, "grafana-loki", res["name"])
	require.Equal(t, loki.Name, res["alias"])
	require.Equal(t, "http:loki:3100", res["url"])
	require.Equal(t, "{job=\"telemetry-fluent-bit\"}", res["labels"])
	require.Equal(t, "key1, key2", res["removeKeys"])
	require.Equal(t, "/files/labelmap.json", res["labelMapPath"])
	require.Equal(t, "warn", res["loglevel"])
	require.Equal(t, "json", res["lineformat"])
}
