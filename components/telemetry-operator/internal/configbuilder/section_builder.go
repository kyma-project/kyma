package configbuilder

import (
	"fmt"
	"sort"
	"strings"
)

type configParam struct {
	key   string
	value string
}

type SectionBuilder struct {
	params  []configParam
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
	sb.params = append(sb.params, configParam{key, value})
	return sb
}

func (sb *SectionBuilder) AddIfNotEmpty(key string, value string) *SectionBuilder {
	if value != "" {
		sb.AddConfigParam(key, value)
	}
	return sb
}

func (sb *SectionBuilder) Build() string {
	sort.Slice(sb.params, func(i, j int) bool {
		if sb.params[i].key != sb.params[j].key {
			return sb.params[i].key < sb.params[j].key
		}
		return sb.params[i].value < sb.params[j].value
	})
	indentation := strings.Repeat(" ", 4)
	for _, p := range sb.params {
		sb.builder.WriteString(fmt.Sprintf("%s%s%s%s",
			indentation,
			p.key,
			strings.Repeat(" ", sb.keyLen-len(p.key)+1),
			p.value))
		sb.builder.WriteByte('\n')
	}
	sb.builder.WriteByte('\n')
	return sb.builder.String()
}
