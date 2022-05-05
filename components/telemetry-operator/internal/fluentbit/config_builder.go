package fluentbit

import (
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

type ConfigHeader string

const (
	ParserConfigHeader          ConfigHeader = "[PARSER]"
	MultiLineParserConfigHeader ConfigHeader = "[MULTILINE_PARSER]"
	FilterConfigHeader          ConfigHeader = "[FILTER]"
	OutputConfigHeader          ConfigHeader = "[OUTPUT]"
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
func MergeSectionsConfig(logPipeline *telemetryv1alpha1.LogPipeline) string {
	var sb strings.Builder
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
