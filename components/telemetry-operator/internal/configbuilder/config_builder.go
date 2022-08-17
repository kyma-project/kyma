package configbuilder

import (
	"fmt"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"sort"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type ConfigHeader string

type PipelineConfig struct {
	InputTag          string
	MemoryBufferLimit string
	StorageType       string
	FsBufferLimit     string
}

const (
	ParserConfigHeader      ConfigHeader = "[PARSER]"
	FilterConfigHeader      ConfigHeader = "[FILTER]"
	OutputConfigHeader      ConfigHeader = "[OUTPUT]"
	MatchKey                string       = "match"
	OutputStorageMaxSizeKey string       = "storage.total_limit_size"
	PermanentFilterTemplate string       = `
name                  record_modifier
match                 %s.*
Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}`
	LuaDeDotFilterTemplate string = `
name                  lua
match                 %s.*
script                /fluent-bit/scripts/filter-script.lua
call                  kubernetes_map_keys`
)

// MergeSectionsConfig merges Fluent Bit filters and outputs to a single Fluent Bit configuration.
func MergeSectionsConfig(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (string, error) {
	var sb strings.Builder
	sb.WriteString(CreateRewriteTagFilterSection(pipelineConfig, logPipeline))
	sb.WriteString(BuildConfigSection(FilterConfigHeader, generateFilter(PermanentFilterTemplate, logPipeline.Name)))

	for _, filter := range logPipeline.Spec.Filters {
		section, err := fluentbit.ParseSection(filter.Custom)
		if err != nil {
			return "", err
		}

		section[MatchKey] = generateMatchCondition(logPipeline.Name)

		sb.WriteString(buildConfigSectionFromMap(FilterConfigHeader, section))
	}

	if logPipeline.Spec.Output.HTTP.Host.IsDefined() && logPipeline.Spec.Output.HTTP.Dedot {
		sb.WriteString(BuildConfigSection(FilterConfigHeader, generateFilter(LuaDeDotFilterTemplate, logPipeline.Name)))
	}

	outputSection := CreateOutputSection(logPipeline, pipelineConfig)
	sb.WriteString(outputSection)

	return sb.String(), nil
}

func buildConfigSectionFromMap(header ConfigHeader, section map[string]string) string {
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

// MergeParsersConfig merges Fluent Bit parsers to a single Fluent Bit configuration.
func MergeParsersConfig(logParsers *telemetryv1alpha1.LogParserList) string {
	sort.Slice(logParsers.Items, func(i, j int) bool {
		return logParsers.Items[i].Name < logParsers.Items[j].Name
	})

	var sb strings.Builder
	for _, logParser := range logParsers.Items {
		if logParser.DeletionTimestamp == nil {
			name := fmt.Sprintf("Name %s", logParser.Name)
			parser := fmt.Sprintf("%s\n%s", logParser.Spec.Parser, name)
			sb.WriteString(BuildConfigSection(ParserConfigHeader, parser))
		}
	}
	return sb.String()
}

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

func generateMatchCondition(logPipelineName string) string {
	return fmt.Sprintf("%s.*", logPipelineName)
}

func generateFilter(template string, params ...any) string {
	return fmt.Sprintf(template, params...)
}
