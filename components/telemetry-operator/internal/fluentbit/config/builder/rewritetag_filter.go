package builder

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func createRewriteTagFilter(logPipeline *telemetryv1alpha1.LogPipeline, defaults PipelineDefaults) string {
	var sectionBuilder = NewFilterSectionBuilder().
		AddConfigParam("Name", "rewrite_tag").
		AddConfigParam("Match", fmt.Sprintf("%s.*", defaults.InputTag)).
		AddConfigParam("Emitter_Name", logPipeline.Name).
		AddConfigParam("Emitter_Storage.type", defaults.StorageType).
		AddConfigParam("Emitter_Mem_Buf_Limit", defaults.MemoryBufferLimit)

	containers := logPipeline.Spec.Input.Application.Containers
	if len(containers.Include) > 0 {
		return sectionBuilder.
			AddConfigParam("Rule", fmt.Sprintf("$kubernetes['container_name'] \"^(%s)$\" %s.$TAG true",
				strings.Join(containers.Include, "|"), logPipeline.Name)).
			Build()
	}

	if len(containers.Exclude) > 0 {
		return sectionBuilder.
			AddConfigParam("Rule", fmt.Sprintf("$kubernetes['container_name'] \"^(?!%s$).*\" %s.$TAG true",
				strings.Join(containers.Exclude, "$|"), logPipeline.Name)).
			Build()
	}

	return sectionBuilder.
		AddConfigParam("Rule", fmt.Sprintf("$log \"^.*$\" %s.$TAG true", logPipeline.Name)).
		Build()
}
