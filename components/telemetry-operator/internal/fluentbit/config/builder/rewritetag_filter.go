package builder

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func systemNamespaces() []string {
	return []string{"kyma-system", "kyma-integration", "kube-system", "istio-system", "compass-system"}
}

func getEmitterPostfixByOutput(output *telemetryv1alpha1.Output) string {
	if output.IsHTTPDefined() {
		return "http"
	}

	if output.IsLokiDefined() {
		return "grafana-loki"
	}

	if !output.IsCustomDefined() {
		return ""
	}

	customOutputParams := parseMultiline(output.Custom)
	postfix := customOutputParams.GetByKey("name")

	if postfix == nil {
		return ""
	}

	return postfix.Value
}

// CreateRewriteTagFilter creates the Fluent Bit Rewrite Tag Filter section
func createRewriteTagFilterSection(logPipeline *telemetryv1alpha1.LogPipeline, defaults PipelineDefaults) string {
	emitterName := logPipeline.Name
	output := &logPipeline.Spec.Output
	emitterPostfix := getEmitterPostfixByOutput(output)

	if emitterPostfix != "" {
		emitterName += ("-" + emitterPostfix)
	}


	var sectionBuilder = NewFilterSectionBuilder().
		AddConfigParam("Name", "rewrite_tag").
		AddConfigParam("Match", fmt.Sprintf("%s.*", defaults.InputTag)).
		AddConfigParam("Emitter_Name", emitterName).
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
