package builder

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"strings"
)

func createNamespaceGrepFilter(logPipeline *telemetryv1alpha1.LogPipeline, defaults PipelineDefaults) string {
	var sectionBuilder = NewFilterSectionBuilder().
		AddConfigParam("Name", "grep").
		AddConfigParam("Match", fmt.Sprintf("%s.*", defaults.InputTag))

	namespaces := logPipeline.Spec.Input.Application.Namespaces
	if len(namespaces.Include) > 0 {
		sectionBuilder.AddConfigParam("Regex",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\"",
				strings.Join(namespaces.Include, "|")))
		return sectionBuilder.Build()
	}

	if len(namespaces.Exclude) > 0 {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!%s$).*\"",
				strings.Join(namespaces.Exclude, "$|")))
		return sectionBuilder.Build()
	}

	if namespaces.System {
		sectionBuilder.AddConfigParam("Rule", fmt.Sprintf("$log \"^.*$\""))
	} else {
		sectionBuilder.AddConfigParam("Rule",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(?!%s$).*\"",
				strings.Join(systemNamespaces(), "$|")))
	}
	return sectionBuilder.Build()
}
