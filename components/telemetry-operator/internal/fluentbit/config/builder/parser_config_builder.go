package builder

import (
	"fmt"
	"sort"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

// BuildFluentBitParsersConfig merges Fluent Bit parsers to a single Fluent Bit configuration.
func BuildFluentBitParsersConfig(logParsers *telemetryv1alpha1.LogParserList) string {
	sort.Slice(logParsers.Items, func(i, j int) bool {
		return logParsers.Items[i].Name < logParsers.Items[j].Name
	})

	var sb strings.Builder
	for _, logParser := range logParsers.Items {
		if logParser.DeletionTimestamp.IsZero() {
			sb.WriteString(createParserConfig(logParser.Name, logParser.Spec.Parser))
		}
	}
	return sb.String()
}

func createParserConfig(name, content string) string {
	var sb strings.Builder
	sb.WriteString("[PARSER]\n")
	sb.WriteString("    " + fmt.Sprintf("Name %s\n", name))
	for _, line := range strings.Split(content, "\n") {
		if len(strings.TrimSpace(line)) > 0 { // Skip empty lines to do not break rendering in yaml output
			sb.WriteString("    " + strings.TrimSpace(line) + "\n") // 4 indentations
		}
	}
	sb.WriteByte('\n')
	return sb.String()
}
