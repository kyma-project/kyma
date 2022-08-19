package fluentbit

import (
	"testing"

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
	actual := BuildConfigSection(parserConfigHeader, content)

	require.Equal(t, expected, actual, "Fluent Bit Config Build is invalid")
}

func TestBuildSectionWithWrongIndentation(t *testing.T) {
	expected := `[PARSER]
    Name   dummy_test
    Format   regex
    Regex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$

`
	content := "Name   dummy_test   \n  Format   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
	actual := BuildConfigSection(parserConfigHeader, content)

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
	actual := buildConfigSectionFromMap(filterConfigHeader, content)

	require.Equal(t, expected, actual, "Fluent Bit config Build from Map is invalid")

}

func TestGeneratePermanentFilter(t *testing.T) {
	expected := `
name                  record_modifier
match                 foo.*
Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}`

	actual := generateFilter(permanentFilterTemplate, "foo")
	require.Equal(t, expected, actual, "Fluent Bit Permanent parser config is invalid")
}

func TestGenerateLuaFilter(t *testing.T) {
	expected := `
name                  lua
match                 foo.*
script                /fluent-bit/scripts/filter-script.lua
call                  kubernetes_map_keys`

	actual := generateFilter(luaDeDotFilterTemplate, "foo")
	require.Equal(t, expected, actual, "Fluent Bit lua parser config is invalid")
}

func TestGenerateApplicationFilterDropAll(t *testing.T) {
	expected := `[FILTER]
    name                  nest
    match                 test-logpipeline.*
    operation             lift
    nested_under          kubernetes
    add_prefix            __k8s__

[FILTER]
    name                  record_modifier
    match                 test-logpipeline.*
    remove_key            __k8s__annotations

[FILTER]
    name                  record_modifier
    match                 test-logpipeline.*
    remove_key            __k8s__labels

[FILTER]
    name                  nest
    match                 test-logpipeline.*
    operation             nest
    wildcard              __k8s__*
    nest_under            kubernetes
    remove_prefix         __k8s__

`

	actual := createKubernetesMetadataFilter("test-logpipeline", false, true)
	require.Equal(t, expected, actual, "Fluent Bit application filters are invalid")
}

func TestGenerateApplicationFilterKeepAll(t *testing.T) {
	expected := ""

	actual := createKubernetesMetadataFilter("test-logpipeline", true, false)
	require.Equal(t, expected, actual, "Fluent Bit application filters are invalid")
}

func TestFilter(t *testing.T) {
	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    match foo.*
    name grep
    regex log aa

[FILTER]
    name                  lua
    match                 foo.*
    script                /fluent-bit/scripts/filter-script.lua
    call                  kubernetes_map_keys

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

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false,
				},
			},
			Filters: []telemetryv1alpha1.Filter{
				{
					Custom: `
						name grep
						regex log aa
					`,
				},
			},
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: true,
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
    Rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" foo.$TAG true

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
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false,
				},
			},
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
    Rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    name                  lua
    match                 foo.*
    script                /fluent-bit/scripts/filter-script.lua
    call                  kubernetes_map_keys

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
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false,
				},
			},
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
    Rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    name                  lua
    match                 foo.*
    script                /fluent-bit/scripts/filter-script.lua
    call                  kubernetes_map_keys

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
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false,
				},
			},
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
	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" foo.$TAG true

[FILTER]
    name                  record_modifier
    match                 foo.*
    Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}

[OUTPUT]
    alias foo
    labelMapPath /fluent-bit/etc/loki-labelmap.json
    labels {job="telemetry-fluent-bit"}
    lineformat json
    loglevel warn
    match foo.*
    name grafana-loki
    removeKeys key1, key2
    storage.total_limit_size 1G
    url http:loki:3100

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "lokiNs",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false,
				},
			},
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

	actual, err := MergeSectionsConfig(logPipeline, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
