package builder

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func systemNamespaces() []string {
	return []string{"kyma-system", "kyma-integration", "kube-system", "istio-system", "compass-system"}
}

func createNamespaceGrepFilter(logPipeline *telemetryv1alpha1.LogPipeline, defaults PipelineDefaults) string {
	namespaces := logPipeline.Spec.Input.Application.Namespaces
	if namespaces.System {
		return ""
	}

	var sectionBuilder = NewFilterSectionBuilder().
		AddConfigParam("Name", "grep").
		AddConfigParam("Match", fmt.Sprintf("%s.*", defaults.InputTag))

	if len(namespaces.Include) > 0 {
		sectionBuilder.AddConfigParam("Regex",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\"",
				strings.Join(namespaces.Include, "|")))
		return sectionBuilder.Build()
	}

	if len(namespaces.Exclude) > 0 {
		sectionBuilder.AddConfigParam("Exclude",
			fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\"",
				strings.Join(namespaces.Exclude, "|")))
		return sectionBuilder.Build()
	}

	return sectionBuilder.AddConfigParam("Exclude",
		fmt.Sprintf("$kubernetes['namespace_name'] \"^(%s)$\"",
			strings.Join(systemNamespaces(), "|"))).Build()
}
