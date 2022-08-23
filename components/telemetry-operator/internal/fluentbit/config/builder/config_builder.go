package builder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type PipelineConfig struct {
	InputTag          string
	MemoryBufferLimit string
	StorageType       string
	FsBufferLimit     string
}

// MergeSectionsConfig merges Fluent Bit filters and outputs to a single Fluent Bit configuration.
func MergeSectionsConfig(pipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (string, error) {
	err := validateOutput(pipeline)
	if err != nil {
		return "", err
	}

	err = validateCustomSections(pipeline)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(createRewriteTagFilterSection(pipeline, pipelineConfig))
	sb.WriteString(createRecordModifierFilter(pipeline))
	sb.WriteString(createCustomFilters(pipeline))
	sb.WriteString(createKubernetesMetadataFilter(pipeline))
	sb.WriteString(createLuaDedotFilter(pipeline))
	sb.WriteString(createOutputSection(pipeline, pipelineConfig))

	return sb.String(), nil
}

func createRecordModifierFilter(pipeline *telemetryv1alpha1.LogPipeline) string {
	return NewFilterSectionBuilder().
		AddConfigParam("name", "record_modifier").
		AddConfigParam("match", fmt.Sprintf("%s.*", pipeline.Name)).
		AddConfigParam("record", "cluster_identifier ${KUBERNETES_SERVICE_HOST}").
		Build()
}

func createLuaDedotFilter(logPipeline *telemetryv1alpha1.LogPipeline) string {
	output := logPipeline.Spec.Output
	if !output.HTTPDefined() || !output.HTTP.Dedot {
		return ""
	}

	return NewFilterSectionBuilder().
		AddConfigParam("name", "lua").
		AddConfigParam("match", fmt.Sprintf("%s.*", logPipeline.Name)).
		AddConfigParam("script", "/fluent-bit/scripts/filter-script.lua").
		AddConfigParam("call", "kubernetes_map_keys").
		Build()
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
			sb.WriteString(buildConfigSection("[PARSER]", parser))
		}
	}
	return sb.String()
}

func buildConfigSection(header string, content string) string {
	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteByte('\n')
	for _, line := range strings.Split(content, "\n") {
		if len(strings.TrimSpace(line)) > 0 { // Skip empty lines to do not break rendering in yaml output
			sb.WriteString("    " + strings.TrimSpace(line) + "\n") // 4 indentations
		}
	}
	sb.WriteByte('\n')

	return sb.String()
}

func validateCustomSections(pipeline *telemetryv1alpha1.LogPipeline) error {
	customOutput := pipeline.Spec.Output.Custom
	if customOutput != "" {
		_, err := config.ParseCustomSection(customOutput)
		if err != nil {
			return err
		}
	}

	for _, filter := range pipeline.Spec.Filters {
		_, err := config.ParseCustomSection(filter.Custom)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateOutput(pipeline *telemetryv1alpha1.LogPipeline) error {
	if !pipeline.Spec.Output.AnyDefined() {
		return fmt.Errorf("output plugin not defined")
	}
	return nil
}
