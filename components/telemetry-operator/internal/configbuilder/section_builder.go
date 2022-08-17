package configbuilder

import (
	"fmt"
	"strings"
)

type configParam struct {
	key   string
	value string
}

type SectionBuilder struct {
	indentation string
	valueTab    string
	builder     strings.Builder
}

func NewSectionBuilder() *SectionBuilder {
	return &SectionBuilder{
		indentation: strings.Repeat(" ", 4),
		valueTab:    strings.Repeat(" ", 22),
	}
}

func (sb *SectionBuilder) CreateFilterSection() *SectionBuilder {
	sb.builder.WriteString("[FILTER]")
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) CreateOutputSection() *SectionBuilder {
	sb.builder.WriteString("[OUTPUT]")
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) AddConfigParam(key string, value string) *SectionBuilder {
	sb.builder.WriteString(fmt.Sprintf("%s%s%s%s",
		sb.indentation,
		key,
		sb.valueTab[:len(sb.valueTab)-len(key)],
		value))
	sb.builder.WriteByte('\n')
	return sb
}

func (sb *SectionBuilder) AddIfNotEmpty(key string, value string) *SectionBuilder {
	if value != "" {
		sb.AddConfigParam(key, value)
	}
	return sb
}

func (sb *SectionBuilder) String() string {
	sb.builder.WriteByte('\n')
	return sb.builder.String()
}
