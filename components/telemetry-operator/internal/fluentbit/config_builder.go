package fluentbit

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/secret"
	"sort"
	"strings"
)

type ConfigHeader string

type PipelineConfig struct {
	InputTag          string
	MemoryBufferLimit string
	StorageType       string
	FsBufferLimit     string
}

const (
	ParserConfigHeader      ConfigHeader = "[PARSER]"
	FilterConfigHeader      ConfigHeader = "[FILTER]"
	OutputConfigHeader      ConfigHeader = "[OUTPUT]"
	MatchKey                string       = "match"
	OutputStorageMaxSizeKey string       = "storage.total_limit_size"
	EmitterTemplate         string       = `
name                  rewrite_tag
match                 %s.*
Rule                  $log "^.*$" %s.$TAG true
Emitter_Name          %s
Emitter_Storage.type  %s
Emitter_Mem_Buf_Limit %s`
	PermanentFilterTemplate string = `
name                  record_modifier
match                 %s.*
Record                cluster_identifier ${KUBERNETES_SERVICE_HOST}`
)

func BuildConfigSection(header ConfigHeader, content string) string {
	var sb strings.Builder
	sb.WriteString(string(header))
	sb.WriteByte('\n')
	for _, line := range strings.Split(content, "\n") {
		if len(strings.TrimSpace(line)) > 0 { // Skip empty lines to do not break rendering in yaml output
			sb.WriteString("    " + strings.TrimSpace(line) + "\n") // 4 indentations
		}
	}
	sb.WriteByte('\n')

	return sb.String()
}

func BuildConfigSectionFromMap(header ConfigHeader, section map[string]string) string {
	// Sort maps for idempotent results
	keys := make([]string, 0, len(section))
	for k := range section {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(string(header))
	sb.WriteByte('\n')
	for _, key := range keys {
		sb.WriteString("    " + key + " " + section[key] + "\n") // 4 indentations
	}
	sb.WriteByte('\n')

	return sb.String()
}

// MergeSectionsConfig merges Fluent Bit filters and outputs to a single Fluent Bit configuration.
func MergeSectionsConfig(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString(BuildConfigSection(FilterConfigHeader, generateEmitter(pipelineConfig, logPipeline.Name)))

	sb.WriteString(BuildConfigSection(FilterConfigHeader, generatePermanentFilter(logPipeline.Name)))

	for _, filter := range logPipeline.Spec.Filters {
		section, err := ParseSection(filter.Custom)
		if err != nil {
			return "", err
		}

		section[MatchKey] = generateMatchCondition(logPipeline.Name)

		sb.WriteString(BuildConfigSectionFromMap(FilterConfigHeader, section))
	}

	outputSection, err := generateOutputSection(logPipeline, pipelineConfig)
	if err != nil {
		return "", err
	}
	sb.WriteString(BuildConfigSectionFromMap(OutputConfigHeader, outputSection))

	return sb.String(), nil
}

func generateOutputSection(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (map[string]string, error) {
	if len(logPipeline.Spec.Output.Custom) > 0 {
		return generateCustomOutput(logPipeline, pipelineConfig)
	}

	// An HTTP output needs at least a host attribute
	if logPipeline.Spec.Output.HTTP.Host.IsDefined() {
		return generateHTTPOutput(logPipeline, pipelineConfig)
	}

	// A LokiOutput needs to have atleast url
	if logPipeline.Spec.Output.Loki.URL.IsDefined() {
		return generateLokiOutPut(logPipeline, pipelineConfig)
	}

	return nil, fmt.Errorf("output plugin not defined")
}

func getLokiDefaults() map[string]string {
	result := make(map[string]string)
	result["labelMapPath"] = "/files/labelmap.json"
	result["loglevel"] = "warn"
	result["lineformat"] = "json"
	return result
}
func generateLokiOutPut(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (map[string]string, error) {
	lokiConfig := logPipeline.Spec.Output.Loki
	result := getLokiDefaults()
	result[MatchKey] = generateMatchCondition(logPipeline.Name)
	result[OutputStorageMaxSizeKey] = pipelineConfig.FsBufferLimit
	result["name"] = "grafana-loki"
	result["alias"] = logPipeline.Name
	result["url"] = lokiConfig.URL.Value
	if len(lokiConfig.Labels) != 0 {
		result["labels"] = covertLabelMaptoString(lokiConfig.Labels)
	}
	if len(lokiConfig.RemoveKeys) != 0 {
		str := strings.Join(lokiConfig.RemoveKeys, ", ")
		result["removeKeys"] = str
	}
	return result, nil
}

func covertLabelMaptoString(labels map[string]string) string {
	var labelString []string

	for k, v := range labels {
		l := fmt.Sprintf("%s=\"%s\"", k, v)
		labelString = append(labelString, l)
	}
	return fmt.Sprintf("{%s}", strings.Join(labelString, ", "))
}

func getHTTPOutputDefaults() map[string]string {
	result := map[string]string{
		"name":                     "http",
		"port":                     "443",
		"tls":                      "on",
		"tls.verify":               "on",
		"allow_duplicated_headers": "true",
		"uri":                      "/customindex/kyma",
		"format":                   "json",
	}
	return result
}

func generateHTTPOutput(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (map[string]string, error) {
	result := getHTTPOutputDefaults()
	result[MatchKey] = generateMatchCondition(logPipeline.Name)
	result[OutputStorageMaxSizeKey] = pipelineConfig.FsBufferLimit
	var err error
	if logPipeline.Spec.Output.HTTP.Host.IsDefined() {
		result["host"], err = resolveValue(logPipeline.Spec.Output.HTTP.Host, logPipeline.Name)
		if err != nil {
			return nil, err
		}
	}
	if logPipeline.Spec.Output.HTTP.Password.IsDefined() {
		result["http_passwd"], err = resolveValue(logPipeline.Spec.Output.HTTP.Password, logPipeline.Name)
		if err != nil {
			return nil, err
		}
	}
	if logPipeline.Spec.Output.HTTP.User.IsDefined() {
		result["http_user"], err = resolveValue(logPipeline.Spec.Output.HTTP.User, logPipeline.Name)
		if err != nil {
			return nil, err
		}
	}
	if logPipeline.Spec.Output.HTTP.Port != "" {
		result["port"] = logPipeline.Spec.Output.HTTP.Port
	}
	if logPipeline.Spec.Output.HTTP.URI != "" {
		result["uri"] = logPipeline.Spec.Output.HTTP.URI
	}
	if logPipeline.Spec.Output.HTTP.Format != "" {
		result["format"] = logPipeline.Spec.Output.HTTP.Format
	}
	if logPipeline.Spec.Output.HTTP.TLSConfig.Disabled {
		result["tls"] = "off"
	}
	if logPipeline.Spec.Output.HTTP.TLSConfig.SkipCertificateValidation {
		result["tls.verify"] = "off"
	}
	if logPipeline.Spec.Output.HTTP.Compress != "" {
		result["compress"] = logPipeline.Spec.Output.HTTP.Compress
	}

	return result, nil
}

func resolveValue(value telemetryv1alpha1.ValueType, logPipeline string) (string, error) {
	if value.Value != "" {
		return value.Value, nil
	}
	if value.ValueFrom.SecretKey.Name != "" && value.ValueFrom.SecretKey.Key != "" {
		return fmt.Sprintf("${%s}", secret.GenerateVariableName(value.ValueFrom.SecretKey, logPipeline)), nil
	}
	return "", fmt.Errorf("value not defined")
}

func generateCustomOutput(logPipeline *telemetryv1alpha1.LogPipeline, pipelineConfig PipelineConfig) (map[string]string, error) {
	section, err := ParseSection(logPipeline.Spec.Output.Custom)
	if err != nil {
		return nil, err
	}

	section[MatchKey] = generateMatchCondition(logPipeline.Name)
	section[OutputStorageMaxSizeKey] = pipelineConfig.FsBufferLimit

	return section, nil
}

func generateMatchCondition(logPipelineName string) string {
	return fmt.Sprintf("%s.*", logPipelineName)
}

// MergeParsersConfig merges Fluent Bit parsers to a single Fluent Bit configuration.
func MergeParsersConfig(logParsers *telemetryv1alpha1.LogParserList) string {
	var sb strings.Builder
	for _, logParser := range logParsers.Items {
		var parser string
		if logParser.DeletionTimestamp == nil {
			name := fmt.Sprintf("Name %s", logParser.Name)
			parser = fmt.Sprintf("%s\n%s", logParser.Spec.Parser, name)
			sb.WriteString(BuildConfigSection(ParserConfigHeader, parser))
		}
	}
	return sb.String()
}

func generateEmitter(config PipelineConfig, logPipelineName string) string {
	return fmt.Sprintf(EmitterTemplate, config.InputTag, logPipelineName, logPipelineName, config.StorageType, config.MemoryBufferLimit)
}

func generatePermanentFilter(logPipelineName string) string {
	return fmt.Sprintf(PermanentFilterTemplate, logPipelineName)
}
