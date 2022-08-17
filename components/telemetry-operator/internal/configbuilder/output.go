package configbuilder

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"strings"
)

// CreateOutputSection creates the Fluent Bit Output section
func CreateOutputSection(pipeline *telemetryv1alpha1.LogPipeline, config PipelineConfig) string {
	output := &pipeline.Spec.Output
	if len(output.Custom) > 0 {
		return generateCustomOutput(output, config.FsBufferLimit, pipeline.Name)
	} else if output.HTTP.Host.IsDefined() {
		return generateHTTPOutput(&output.HTTP, config.FsBufferLimit, pipeline.Name)
	} else if output.Loki.URL.IsDefined() {
		return generateLokiOutput(&output.Loki, config.FsBufferLimit, pipeline.Name)
	}
	return ""
}

func generateCustomOutput(output *telemetryv1alpha1.Output, fsBufferLimit string, name string) string {
	sb := NewSectionBuilder()
	sb.CreateOutputSection()
	customOutputParams := parseMultiline(output.Custom)
	for _, p := range customOutputParams {
		sb.AddConfigParam(p.key, p.value)
	}
	sb.AddConfigParam(MatchKey, generateMatchCondition(name))
	sb.AddConfigParam(OutputStorageMaxSizeKey, fsBufferLimit)
	return sb.String()
}

func generateHTTPOutput(httpOutput *telemetryv1alpha1.HTTPOutput, fsBufferLimit string, name string) string {
	sb := NewSectionBuilder()
	sb.CreateOutputSection()
	sb.AddConfigParam("name", "http")
	sb.AddConfigParam("port", "443")
	sb.AddConfigParam("tls", "on")
	sb.AddConfigParam("allow_duplicated_headers", "true")
	sb.AddConfigParam("format", "json")
	sb.AddConfigParam(MatchKey, generateMatchCondition(name))
	sb.AddConfigParam(OutputStorageMaxSizeKey, fsBufferLimit)
	if httpOutput.Host.IsDefined() {
		value := resolveValue(httpOutput.Host, name)
		sb.AddConfigParam("host", value)
	}
	if httpOutput.Password.IsDefined() {
		value := resolveValue(httpOutput.Password, name)
		sb.AddConfigParam("http_passwd", value)
	}
	if httpOutput.User.IsDefined() {
		value := resolveValue(httpOutput.User, name)
		sb.AddConfigParam("http_user", value)
	}
	sb.AddIfNotEmpty("port", httpOutput.Port)
	sb.AddIfNotEmpty("uri", httpOutput.URI)
	sb.AddIfNotEmpty("format", httpOutput.Format)
	sb.AddIfNotEmpty("compress", httpOutput.Compress)
	if httpOutput.TLSConfig.Disabled {
		sb.AddConfigParam("tls", "off")
	}
	if httpOutput.TLSConfig.SkipCertificateValidation {
		sb.AddConfigParam("tls.verify", "off")
	}
	return sb.String()
}

func generateLokiOutput(lokiOutput *telemetryv1alpha1.LokiOutput, fsBufferLimit string, name string) string {
	sb := NewSectionBuilder()
	sb.CreateOutputSection()
	sb.AddConfigParam("labelMapPath", "/fluent-bit/etc/loki-labelmap.json")
	sb.AddConfigParam("loglevel", "warn")
	sb.AddConfigParam("lineformat", "json")
	sb.AddConfigParam(MatchKey, generateMatchCondition(name))
	sb.AddConfigParam(OutputStorageMaxSizeKey, fsBufferLimit)
	sb.AddConfigParam("name", "grafana-loki")
	sb.AddConfigParam("alias", name)
	sb.AddConfigParam("url", resolveValue(lokiOutput.URL, name))
	if len(lokiOutput.Labels) != 0 {
		value := concatenateLabels(lokiOutput.Labels)
		sb.AddConfigParam("labels", value)
	}
	if len(lokiOutput.RemoveKeys) != 0 {
		str := strings.Join(lokiOutput.RemoveKeys, ", ")
		sb.AddConfigParam("removeKeys", str)
	}
	return sb.String()
}

func parseMultiline(section string) []configParam {
	var result []configParam
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		result = append(result, configParam{strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)})
	}
	return result
}

func concatenateLabels(labels map[string]string) string {
	var labelString []string
	for k, v := range labels {
		labelString = append(labelString, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	// {key="value", key="value", key="value", key="value"}
	return fmt.Sprintf("{%s}", strings.Join(labelString, ", "))
}

func resolveValue(value telemetryv1alpha1.ValueType, logPipeline string) string {
	if value.Value != "" {
		return value.Value
	}
	if value.ValueFrom.IsSecretRef() {
		return fmt.Sprintf("${%s}", generateVariableName(value.ValueFrom.SecretKey, logPipeline))
	}
	return ""
}

func generateVariableName(secretRef telemetryv1alpha1.SecretKeyRef, pipelineName string) string {
	result := fmt.Sprintf("%s_%s_%s_%s", pipelineName, secretRef.Namespace, secretRef.Name, secretRef.Key)
	result = strings.ToUpper(result)
	result = strings.Replace(result, ".", "_", -1)
	result = strings.Replace(result, "-", "_", -1)
	return result
}
