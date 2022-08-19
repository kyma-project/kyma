package configbuilder

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sort"
	"strings"
)

func createCustomFilters(pipeline *telemetryv1alpha1.LogPipeline) string {
	var filters []string

	for _, filter := range pipeline.Spec.Filters {
		builder := NewFilterSectionBuilder()
		customFilterParams := parseMultiline(filter.Custom)
		for _, p := range customFilterParams {
			builder.AddConfigParam(p.key, p.value)
		}
		builder.AddConfigParam("match", fmt.Sprintf("%s.*", pipeline.Name))
		filters = append(filters, builder.Build())
	}

	return strings.Join(filters, "")
}

func buildConfigSectionFromMap(header string, section map[string]string) string {
	// Sort maps for idempotent results
	keys := make([]string, 0, len(section))
	for k := range section {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteByte('\n')
	for _, key := range keys {
		sb.WriteString("    " + key + " " + section[key] + "\n") // 4 indentations
	}
	sb.WriteByte('\n')

	return sb.String()
}
