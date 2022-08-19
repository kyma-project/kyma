package configbuilder

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func createRewriteTagFilterSection(logPipeline *telemetryv1alpha1.LogPipeline, config PipelineConfig) string {
	var builder = NewFilterSectionBuilder().
		AddConfigParam("Name", "rewrite_tag").
		AddConfigParam("Match", fmt.Sprintf("%s.*", config.InputTag)).
		AddConfigParam("Emitter_Name", logPipeline.Name).
		AddConfigParam("Emitter_Storage.type", config.StorageType).
		AddConfigParam("Emitter_Mem_Buf_Limit", config.MemoryBufferLimit)

	if !logPipeline.Spec.Input.Application.HasSelectors() {
		if logPipeline.Spec.Input.Application.IncludeSystemNamespaces {
			builder.AddConfigParam("Rule", fmt.Sprintf("$log \"^.*$\" %s.$TAG true", logPipeline.Name))
		} else {
			builder.AddConfigParam("Rule",
				fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*\" %s.$TAG true", logPipeline.Name))
		}
		return builder.Build()
	}

	if len(logPipeline.Spec.Input.Application.Namespaces) > 0 {
		builder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.Namespaces, "|"), logPipeline.Name))
	}

	if len(logPipeline.Spec.Input.Application.ExcludeNamespaces) > 0 {
		builder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.ExcludeNamespaces, "$|"), logPipeline.Name))
	}

	if len(logPipeline.Spec.Input.Application.Containers) > 0 {
		builder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['container_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.Containers, "|"), logPipeline.Name))
	}

	if len(logPipeline.Spec.Input.Application.ExcludeContainers) > 0 {
		builder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['container_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(logPipeline.Spec.Input.Application.ExcludeContainers, "$|"), logPipeline.Name))
	}

	return builder.Build()
}
