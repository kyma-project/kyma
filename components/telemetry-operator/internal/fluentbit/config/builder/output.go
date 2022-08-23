package builder

import (
	"fmt"
	"sort"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"
)

func createOutputSection(pipeline *telemetryv1alpha1.LogPipeline, config PipelineConfig) string {
	output := &pipeline.Spec.Output
	if output.CustomDefined() {
		return generateCustomOutput(output, config.FsBufferLimit, pipeline.Name)
	}

	if output.HTTPDefined() {
		return generateHTTPOutput(output.HTTP, config.FsBufferLimit, pipeline.Name)
	}

	if output.LokiDefined() {
		return generateLokiOutput(output.Loki, config.FsBufferLimit, pipeline.Name)
	}

	return ""
}

func generateCustomOutput(output *telemetryv1alpha1.Output, fsBufferLimit string, name string) string {
	sb := NewOutputSectionBuilder()
	customOutputParams := parseMultiline(output.Custom)
	for _, p := range customOutputParams {
		sb.AddConfigParam(p.Key, p.Value)
	}
	sb.AddConfigParam("match", fmt.Sprintf("%s.*", name))
	sb.AddConfigParam("storage.total_limit_size", fsBufferLimit)
	return sb.Build()
}

func generateHTTPOutput(httpOutput *telemetryv1alpha1.HTTPOutput, fsBufferLimit string, name string) string {
	sb := NewOutputSectionBuilder()
	sb.AddConfigParam("name", "http")
	sb.AddConfigParam("port", "443")
	sb.AddConfigParam("allow_duplicated_headers", "true")
	sb.AddConfigParam("format", "json")
	sb.AddConfigParam("match", fmt.Sprintf("%s.*", name))
	sb.AddConfigParam("storage.total_limit_size", fsBufferLimit)
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
	} else {
		sb.AddConfigParam("tls", "on")
	}
	if httpOutput.TLSConfig.SkipCertificateValidation {
		sb.AddConfigParam("tls.verify", "off")
	} else {
		sb.AddConfigParam("tls.verify", "on")
	}
	return sb.Build()
}

func generateLokiOutput(lokiOutput *telemetryv1alpha1.LokiOutput, fsBufferLimit string, name string) string {
	sb := NewOutputSectionBuilder()
	sb.AddConfigParam("labelMapPath", "/fluent-bit/etc/loki-labelmap.json")
	sb.AddConfigParam("loglevel", "warn")
	sb.AddConfigParam("lineformat", "json")
	sb.AddConfigParam("match", fmt.Sprintf("%s.*", name))
	sb.AddConfigParam("storage.total_limit_size", fsBufferLimit)
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
	return sb.Build()
}

func concatenateLabels(labels map[string]string) string {
	var labelsSlice []string
	for k, v := range labels {
		labelsSlice = append(labelsSlice, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	sort.Strings(labelsSlice)
	return fmt.Sprintf("{%s}", strings.Join(labelsSlice, ", "))
}

func resolveValue(value telemetryv1alpha1.ValueType, logPipeline string) string {
	if value.Value != "" {
		return value.Value
	}
	if value.ValueFrom.IsSecretRef() {
		return fmt.Sprintf("${%s}", envvar.GenerateName(logPipeline, value.ValueFrom.SecretKey))
	}
	return ""
}
