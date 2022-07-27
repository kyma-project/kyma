package fluentbit

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type SectionBuilder struct {
	indentation string
	valueTab    string
	builder     strings.Builder
}

func NewSectionBuilder() *SectionBuilder {
	return &SectionBuilder{
		indentation: strings.Repeat(" ", 4),
		valueTab:    strings.Repeat(" ", 22),
	}
}

func (sb *SectionBuilder) CreateFilterSection() *SectionBuilder {
	sb.builder.WriteString("[FILTER]")
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) AddConfigurationParameter(key string, value string) *SectionBuilder {
	sb.builder.WriteString(fmt.Sprintf("%s%s%s%s",
		sb.indentation,
		key,
		sb.valueTab[:len(sb.valueTab)-len(key)],
		value))
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) ToString() string {
	sb.builder.WriteByte('\n')
	return sb.builder.String()
}

// CreateEmitter creates the Fluent Bit Rewrite Tag Filter section
func CreateEmitter(config PipelineConfig, logPipeline *telemetryv1alpha1.LogPipeline) string {

	var sectionBuilder = NewSectionBuilder().
		CreateFilterSection().
		AddConfigurationParameter("Name", "rewrite_tag").
		AddConfigurationParameter("Match", fmt.Sprintf("%s.*", config.InputTag)).
		AddConfigurationParameter("Emitter_Name", logPipeline.Name).
		AddConfigurationParameter("Emitter_Storage.type", config.StorageType).
		AddConfigurationParameter("Emitter_Mem_Buf_Limit", config.MemoryBufferLimit)

	if !logPipeline.Spec.Input.Application.HasSelectors() {
		sectionBuilder.AddConfigurationParameter("Rule", fmt.Sprintf("$log \"^.*$\" %s.$TAG true", logPipeline.Name))
		return sectionBuilder.ToString()
	}

	if len(logPipeline.Spec.Input.Application.Namespaces) > 0 {
		sectionBuilder.AddConfigurationParameter("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.Namespaces, "|"), logPipeline.Name))
	}

	if len(logPipeline.Spec.Input.Application.ExcludeNamespaces) > 0 {
		sectionBuilder.AddConfigurationParameter("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.ExcludeNamespaces, "$|"), logPipeline.Name))
	}

	if len(logPipeline.Spec.Input.Application.Containers) > 0 {
		sectionBuilder.AddConfigurationParameter("Rule",
			fmt.Sprintf("$kubernetes['container_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.Containers, "|"), logPipeline.Name))
	}

	if len(logPipeline.Spec.Input.Application.ExcludeContainers) > 0 {
		sectionBuilder.AddConfigurationParameter("Rule",
			fmt.Sprintf("$kubernetes['container_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.ExcludeContainers, "$|"), logPipeline.Name))
	}

	return sectionBuilder.ToString()
}
