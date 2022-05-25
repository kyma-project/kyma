package fluentbit

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

type ConfigHeader string

type EmitterConfig struct {
	InputTag    string
	BufferLimit string
	StorageType string
}

const (
	ParserConfigHeader          ConfigHeader = "[PARSER]"
	MultiLineParserConfigHeader ConfigHeader = "[MULTILINE_PARSER]"
	FilterConfigHeader          ConfigHeader = "[FILTER]"
	OutputConfigHeader          ConfigHeader = "[OUTPUT]"
	EmitterTemplate             string       = `
name                  rewrite_tag
match                 %s.*
Rule                  $log "^.*$" %s$TAG true
Emitter_Name          %s
Emitter_Storage.type  %s
Emitter_Mem_Buf_Limit %s`
)

func BuildConfigSection(header ConfigHeader, content string) string {
	var sb strings.Builder
	sb.WriteString(string(header))
	sb.WriteByte('\n')
	for _, line := range strings.Split(content, "\n") {
		if len(strings.TrimSpace(line)) > 0 { // Skip empty lines to do not break rendering in yaml output
			sb.WriteString("    " + strings.TrimSpace(line) + "\n") // 4 indentations
		}
	}
	sb.WriteByte('\n')

	return sb.String()
}

// MergeSectionsConfig merges Fluent Bit filters and outputs to a single Fluent Bit configuration.
func MergeSectionsConfig(logPipeline *telemetryv1alpha1.LogPipeline, emitterConfig EmitterConfig) string {
	var sb strings.Builder

	if len(logPipeline.Spec.Outputs) > 0 {
		sb.WriteString(BuildConfigSection(FilterConfigHeader, generateEmitter(emitterConfig, logPipeline)))
	}
	for _, filter := range logPipeline.Spec.Filters {
		sb.WriteString(BuildConfigSection(FilterConfigHeader, filter.Content))
	}
	for _, output := range logPipeline.Spec.Outputs {
		sb.WriteString(BuildConfigSection(OutputConfigHeader, output.Content))
	}
	return sb.String()
}

// MergeParsersConfig merges Fluent Bit parsers and multiLine parsers to a single Fluent Bit configuration.
func MergeParsersConfig(logPipelines *telemetryv1alpha1.LogPipelineList) string {
	var sb strings.Builder
	for _, logPipeline := range logPipelines.Items {
		if logPipeline.DeletionTimestamp == nil {
			for _, parser := range logPipeline.Spec.Parsers {
				sb.WriteString(BuildConfigSection(ParserConfigHeader, parser.Content))
			}
			for _, multiLineParser := range logPipeline.Spec.MultiLineParsers {
				sb.WriteString(BuildConfigSection(MultiLineParserConfigHeader, multiLineParser.Content))
			}
		}
	}
	return sb.String()
}

func generateEmitter(emitterConfig EmitterConfig, logPipeline *telemetryv1alpha1.LogPipeline) string {
	return fmt.Sprintf(EmitterTemplate, emitterConfig.InputTag, logPipeline.Name, logPipeline.Name, emitterConfig.StorageType, emitterConfig.BufferLimit)
}
