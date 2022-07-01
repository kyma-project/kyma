package fluentbit

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/log-pipelines/v1alpha1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSection(t *testing.T) {
	expected := `[PARSER]
    Name   dummy_test
    Format   regex
    Regex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$

`
	content := "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
	actual := BuildConfigSection(ParserConfigHeader, content)

	assert.Equal(t, expected, actual, "Fluent Bit Config Build is invalid")
}

func TestBuildSectionWithWrongIndentation(t *testing.T) {
	expected := `[PARSER]
    Name   dummy_test
    Format   regex
    Regex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$

`
	content := "Name   dummy_test   \n  Format   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$"
	actual := BuildConfigSection(ParserConfigHeader, content)

	assert.Equal(t, expected, actual, "Fluent Bit config indentation has not been fixed")
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
	assert.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}

func TestFilter(t *testing.T) {
	expected := `[FILTER]
    name               grep
    Match              foo.*

`

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
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

	actual, err := MergeSectionsConfig(logPipeline, emitterConfig)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

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
    name               http
    Match              foo.*

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
	emitterConfig := EmitterConfig{
		InputTag:    "kube",
		BufferLimit: "10M",
		StorageType: "filesystem",
	}

	actual, err := MergeSectionsConfig(logPipeline, emitterConfig)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
