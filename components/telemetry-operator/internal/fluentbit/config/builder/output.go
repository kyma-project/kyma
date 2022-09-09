package builder

import (
	"fmt"
	"sort"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"
)

func createOutputSection(pipeline *telemetryv1alpha1.LogPipeline, defaults PipelineDefaults) string {
	output := &pipeline.Spec.Output
	if output.IsCustomDefined() {
		return generateCustomOutput(output, defaults.FsBufferLimit, pipeline.Name)
	}

	if output.IsHTTPDefined() {
		return generateHTTPOutput(output.HTTP, defaults.FsBufferLimit, pipeline.Name)
	}

	if output.IsLokiDefined() {
		return generateLokiOutput(output.Loki, defaults.FsBufferLimit, pipeline.Name)
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
	sb.AddConfigParam("allow_duplicated_headers", "true")
	sb.AddConfigParam("match", fmt.Sprintf("%s.*", name))
	sb.AddConfigParam("alias", fmt.Sprintf("%s-http", name))
	sb.AddConfigParam("storage.total_limit_size", fsBufferLimit)
	sb.AddIfNotEmpty("uri", httpOutput.URI)
	sb.AddIfNotEmpty("compress", httpOutput.Compress)
	sb.AddIfNotEmptyOrDefault("port", httpOutput.Port, "443")
	sb.AddIfNotEmptyOrDefault("format", httpOutput.Format, "json")

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
	tlsEnabled := "on"
	if httpOutput.TLSConfig.Disabled {
		tlsEnabled = "off"
	}
	sb.AddConfigParam("tls", tlsEnabled)
	tlsVerify := "on"
	if httpOutput.TLSConfig.SkipCertificateValidation {
		tlsVerify = "off"
	}
	sb.AddConfigParam("tls.verify", tlsVerify)

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
	sb.AddConfigParam("alias", fmt.Sprintf("%s-grafana-loki", name))
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
	if value.ValueFrom != nil && value.ValueFrom.IsSecretKeyRef() {
		return fmt.Sprintf("${%s}", envvar.GenerateName(logPipeline, *value.ValueFrom.SecretKeyRef))
	}
	return ""
}
