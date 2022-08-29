package builder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"
)

type SectionBuilder struct {
	params  []config.Parameter
	keyLen  int
	builder strings.Builder
}

func NewFilterSectionBuilder() *SectionBuilder {
	sb := SectionBuilder{}
	return sb.createFilterSection()
}

func NewOutputSectionBuilder() *SectionBuilder {
	sb := SectionBuilder{}
	return sb.createOutputSection()
}

func (sb *SectionBuilder) createFilterSection() *SectionBuilder {
	sb.builder.WriteString("[FILTER]")
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) createOutputSection() *SectionBuilder {
	sb.builder.WriteString("[OUTPUT]")
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) AddConfigParam(key string, value string) *SectionBuilder {
	if sb.keyLen < len(key) {
		sb.keyLen = len(key)
	}
	sb.params = append(sb.params, config.Parameter{Key: strings.ToLower(key), Value: value})
	return sb
}

func (sb *SectionBuilder) AddIfNotEmpty(key string, value string) *SectionBuilder {
	if value != "" {
		sb.AddConfigParam(key, value)
	}
	return sb
}

func (sb *SectionBuilder) AddIfNotEmptyOrDefault(key string, value string, defaultValue string) *SectionBuilder {
	if value == "" {
		sb.AddConfigParam(key, defaultValue)
	} else {
		sb.AddConfigParam(key, value)
	}
	return sb
}

func (sb *SectionBuilder) Build() string {
	sort.Slice(sb.params, func(i, j int) bool {
		if sb.params[i].Key != sb.params[j].Key {
			if sb.params[i].Key == "name" {
				return true
			}
			if sb.params[j].Key == "name" {
				return false
			}
			if sb.params[i].Key == "match" {
				return true
			}
			if sb.params[j].Key == "match" {
				return false
			}

			return sb.params[i].Key < sb.params[j].Key
		}
		return sb.params[i].Value < sb.params[j].Value
	})
	indentation := strings.Repeat(" ", 4)
	for _, p := range sb.params {
		sb.builder.WriteString(fmt.Sprintf("%s%s%s%s",
			indentation,
			p.Key,
			strings.Repeat(" ", sb.keyLen-len(p.Key)+1),
			p.Value))
		sb.builder.WriteByte('\n')
	}
	sb.builder.WriteByte('\n')
	return sb.builder.String()
}

func parseMultiline(section string) []config.Parameter {
	var result []config.Parameter
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		param := config.Parameter{Key: strings.ToLower(strings.TrimSpace(key)), Value: strings.TrimSpace(value)}
		result = append(result, param)
	}
	return result
}
