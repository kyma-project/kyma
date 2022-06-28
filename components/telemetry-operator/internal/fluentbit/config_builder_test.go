package fluentbit

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
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
	emitterConfig := EmitterConfig{
		InputTag:    "kube",
		BufferLimit: "10M",
		StorageType: "filesystem",
	}

	expected := `
name                  rewrite_tag
match                 kube.*
Rule                  $log "^.*$" test.$TAG true
Emitter_Name          test
Emitter_Storage.type  filesystem
Emitter_Mem_Buf_Limit 10M`

	actual := generateEmitter(emitterConfig, "test")
	require.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}

func TestFilter(t *testing.T) {
	expected := `[FILTER]
    match foo.*
    name grep

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{
					Content: `
	name               grep	`,
				},
			},
		},
	}
	logPipeline.Name = "foo"
	emitterConfig := EmitterConfig{
		InputTag:    "kube",
		BufferLimit: "10M",
		StorageType: "filesystem",
	}

	actual, err := MergeSectionsConfig(logPipeline, emitterConfig, "1G")
	require.NoError(t, err)
	require.Equal(t, expected, actual)

}

func TestOutput(t *testing.T) {
	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    Rule                  $log "^.*$" foo.$TAG true
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M

[OUTPUT]
    match foo.*
    name http
    storage.total_limit_size 1G

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Outputs: []telemetryv1alpha1.Output{
				{
					Content: `
    name               http`,
				},
			},
		},
	}
	logPipeline.Name = "foo"
	emitterConfig := EmitterConfig{
		InputTag:    "kube",
		BufferLimit: "10M",
		StorageType: "filesystem",
	}

	actual, err := MergeSectionsConfig(logPipeline, emitterConfig, "1G")
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
