package fluentbit

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/log-pipelines/v1alpha1"
	"strings"
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
Rule                  $log "^.*$" %s.$TAG true
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
func MergeSectionsConfig(logPipeline *telemetryv1alpha1.LogPipeline, emitterConfig EmitterConfig) (string, error) {
	var sb strings.Builder

	if len(logPipeline.Spec.Outputs) > 0 {
		sb.WriteString(BuildConfigSection(FilterConfigHeader, generateEmitter(emitterConfig, logPipeline.Name)))
	}
	for _, filter := range logPipeline.Spec.Filters {
		filterSection, err := ensureMatchCondIsValid(filter.Content, logPipeline.Name)
		if err != nil {
			return "", err
		}
		sb.WriteString(BuildConfigSection(FilterConfigHeader, filterSection))
	}
	for _, output := range logPipeline.Spec.Outputs {
		outputSection, err := ensureMatchCondIsValid(output.Content, logPipeline.Name)
		if err != nil {
			return "", err
		}
		sb.WriteString(BuildConfigSection(OutputConfigHeader, outputSection))
	}
	return sb.String(), nil
}

func ensureMatchCondIsValid(content, logPipelineName string) (string, error) {
	section, err := parseSection(content)
	if err != nil {
		return "", err
	}

	matchCond := getMatchCondition(section)
	if matchCond == "" {
		content += fmt.Sprintf("\nMatch              %s.*", logPipelineName)
	}

	return content, nil
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

func generateEmitter(emitterConfig EmitterConfig, logPipelineName string) string {
	return fmt.Sprintf(EmitterTemplate, emitterConfig.InputTag, logPipelineName, logPipelineName, emitterConfig.StorageType, emitterConfig.BufferLimit)
}

func getMatchCondition(section map[string]string) string {
	if matchCond, hasKey := section["match"]; hasKey {
		return matchCond
	}
	return ""
}
