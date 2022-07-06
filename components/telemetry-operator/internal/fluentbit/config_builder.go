package fluentbit

import (
	"fmt"
	"sort"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

type ConfigHeader string

type PipelineConfig struct {
	InputTag          string
	MemoryBufferLimit string
	StorageType       string
	FsBufferLimit     string
}

const (
	ParserConfigHeader          ConfigHeader = "[PARSER]"
	MultiLineParserConfigHeader ConfigHeader = "[MULTILINE_PARSER]"
	FilterConfigHeader          ConfigHeader = "[FILTER]"
	OutputConfigHeader          ConfigHeader = "[OUTPUT]"
	MatchKey                    string       = "match"
	OutputStorageMaxSizeKey     string       = "storage.total_limit_size"
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

func BuildConfigSectionFromMap(header ConfigHeader, section map[string]string) string {
	// Sort maps for idempotent results
	keys := make([]string, 0, len(section))
	for k := range section {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(string(header))
	sb.WriteByte('\n')
	for _, key := range keys {
		sb.WriteString("    " + key + " " + section[key] + "\n") // 4 indentations
	}
	sb.WriteByte('\n')

	return sb.String()
}

// MergeSectionsConfig merges Fluent Bit filters and outputs to a single Fluent Bit configuration.
func MergeSectionsConfig(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (string, error) {
	var sb strings.Builder

	if len(logPipeline.Spec.Output.Custom) > 0 {
		sb.WriteString(BuildConfigSection(FilterConfigHeader, generateEmitter(pipelineConfig, logPipeline.Name)))
	}

	for _, filter := range logPipeline.Spec.Filters {
		section, err := ParseSection(filter.Custom)
		if err != nil {
			return "", err
		}

		section[MatchKey] = getValidMatchCond(section, logPipeline.Name)

		sb.WriteString(BuildConfigSectionFromMap(FilterConfigHeader, section))
	}

	if len(logPipeline.Spec.Output.Custom) > 0 {
		section, err := ParseSection(logPipeline.Spec.Output.Custom)
		if err != nil {
			return "", err
		}

		section[MatchKey] = getValidMatchCond(section, logPipeline.Name)
		section[OutputStorageMaxSizeKey] = pipelineConfig.FsBufferLimit

		sb.WriteString(BuildConfigSectionFromMap(OutputConfigHeader, section))
	}

	return sb.String(), nil
}

func getValidMatchCond(section map[string]string, logPipelineName string) string {
	if matchCond, hasKey := section["match"]; hasKey {
		return matchCond
	}

	return fmt.Sprintf("%s.*", logPipelineName)
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

func generateEmitter(config PipelineConfig, logPipelineName string) string {
	return fmt.Sprintf(EmitterTemplate, config.InputTag, logPipelineName, logPipelineName, config.StorageType, config.MemoryBufferLimit)
}
