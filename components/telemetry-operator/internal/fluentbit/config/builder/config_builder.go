package builder

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type PipelineDefaults struct {
	InputTag          string
	MemoryBufferLimit string
	StorageType       string
	FsBufferLimit     string
}

// BuildFluentBitConfig merges Fluent Bit filters and outputs to a single Fluent Bit configuration.
func BuildFluentBitConfig(pipeline *telemetryv1alpha1.LogPipeline, defaults PipelineDefaults) (string, error) {
	err := validateOutput(pipeline)
	if err != nil {
		return "", err
	}

	err = validateCustomSections(pipeline)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(createRewriteTagFilter(pipeline, defaults))
	sb.WriteString(createNamespaceGrepFilter(pipeline, defaults))
	sb.WriteString(createRecordModifierFilter(pipeline))
	sb.WriteString(createCustomFilters(pipeline))
	sb.WriteString(createKubernetesMetadataFilter(pipeline))
	sb.WriteString(createLuaDedotFilter(pipeline))
	sb.WriteString(createOutputSection(pipeline, defaults))

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
	if !output.IsHTTPDefined() || !output.HTTP.Dedot {
		return ""
	}

	return NewFilterSectionBuilder().
		AddConfigParam("name", "lua").
		AddConfigParam("match", fmt.Sprintf("%s.*", logPipeline.Name)).
		AddConfigParam("script", "/fluent-bit/scripts/filter-script.lua").
		AddConfigParam("call", "kubernetes_map_keys").
		Build()
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
	if !pipeline.Spec.Output.IsAnyDefined() {
		return fmt.Errorf("output plugin not defined")
	}
	return nil
}
