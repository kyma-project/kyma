package builder

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func createCustomFilters(pipeline *telemetryv1alpha1.LogPipeline) string {
	var filters []string

	for _, filter := range pipeline.Spec.Filters {
		builder := NewFilterSectionBuilder()
		customFilterParams := parseMultiline(filter.Custom)
		for _, p := range customFilterParams {
			builder.AddConfigParam(p.Key, p.Value)
		}
		builder.AddConfigParam("match", fmt.Sprintf("%s.*", pipeline.Name))
		filters = append(filters, builder.Build())
	}

	return strings.Join(filters, "")
}
