package configbuilder

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

	actual := generateFilter(PermanentFilterTemplate, "foo")
	require.Equal(t, expected, actual, "Fluent Bit Permanent parser config is invalid")
}

func TestGenerateLuaFilter(t *testing.T) {
	expected := `
name                  lua
match                 foo.*
script                /fluent-bit/scripts/filter-script.lua
call                  kubernetes_map_keys`

	actual := generateFilter(LuaDeDotFilterTemplate, "foo")
	require.Equal(t, expected, actual, "Fluent Bit lua parser config is invalid")
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
