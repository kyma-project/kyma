package fluentbit

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
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
	logPipeline := telemetryv1alpha1.LogPipeline{}
	logPipeline.Name = "test"

	expected := `
name                  rewrite_tag
match                 kube.*
Rule                  $log "^.*$" test.$TAG true
Emitter_Name          test
Emitter_Storage.type  filesystem
Emitter_Mem_Buf_Limit 10M`

	actual := generateEmitter(emitterConfig, &logPipeline)
	assert.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}
