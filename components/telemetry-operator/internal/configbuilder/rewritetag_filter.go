package configbuilder

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func systemNamespaces() []string {
	return []string{"kyma-system", "kyma-integration", "kube-system", "istio-system"}
}

// CreateRewriteTagFilter creates the Fluent Bit Rewrite Tag Filter section
func createRewriteTagFilterSection(logPipeline *telemetryv1alpha1.LogPipeline, config PipelineConfig) string {
	var sectionBuilder = NewFilterSectionBuilder().
		AddConfigParam("Name", "rewrite_tag").
		AddConfigParam("Match", fmt.Sprintf("%s.*", config.InputTag)).
		AddConfigParam("Emitter_Name", logPipeline.Name).
		AddConfigParam("Emitter_Storage.type", config.StorageType).
		AddConfigParam("Emitter_Mem_Buf_Limit", config.MemoryBufferLimit)

	containers := logPipeline.Spec.Input.Application.Containers
	if len(containers.Include) > 0 {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['container_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(containers.Include, "|"), logPipeline.Name))
	}

	if len(containers.Exclude) > 0 {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['container_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(containers.Exclude, "$|"), logPipeline.Name))
	}

	namespaces := logPipeline.Spec.Input.Application.Namespaces
	if len(namespaces.Include) > 0 {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(namespaces.Include, "|"), logPipeline.Name))
		return sectionBuilder.Build()
	}

	if len(namespaces.Exclude) > 0 {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(namespaces.Exclude, "$|"), logPipeline.Name))
		return sectionBuilder.Build()
	}

	if namespaces.System {
		sectionBuilder.AddConfigParam("Rule", fmt.Sprintf("$log \"^.*$\" %s.$TAG true", logPipeline.Name))
	} else {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(systemNamespaces(), "$|"), logPipeline.Name))
	}

	return sectionBuilder.Build()
}
