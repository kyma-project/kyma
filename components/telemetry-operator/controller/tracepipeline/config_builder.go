package tracepipeline

import (
	"fmt"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"
)

type TLSConfig struct {
	Insecure bool `yaml:"insecure"`
}

type SendingQueueConfig struct {
	Enabled   bool `yaml:"enabled"`
	QueueSize int  `yaml:"queue_size"`
}

type RetryOnFailureConfig struct {
	Enabled         bool   `yaml:"enabled"`
	InitialInterval string `yaml:"initial_interval"`
	MaxInterval     string `yaml:"max_interval"`
	MaxElapsedTime  string `yaml:"max_elapsed_time"`
}

type OTLPExporterConfig struct {
	Endpoint       string               `yaml:"endpoint,omitempty"`
	Headers        map[string]string    `yaml:"headers,omitempty"`
	TLS            TLSConfig            `yaml:"tls,omitempty"`
	SendingQueue   SendingQueueConfig   `yaml:"sending_queue,omitempty"`
	RetryOnFailure RetryOnFailureConfig `yaml:"retry_on_failure,omitempty"`
}

type LoggingExporterConfig struct {
	Verbosity string `yaml:"verbosity"`
}

type ExporterConfig struct {
	OTLP     OTLPExporterConfig    `yaml:"otlp,omitempty"`
	OTLPHTTP OTLPExporterConfig    `yaml:"otlphttp,omitempty"`
	Logging  LoggingExporterConfig `yaml:"logging,omitempty"`
}

type ReceiverConfig struct {
	OpenCensus map[string]any `yaml:"opencensus"`
	OTLP       map[string]any `yaml:"otlp"`
}

type BatchProcessorConfig struct {
	SendBatchSize    int    `yaml:"send_batch_size"`
	Timeout          string `yaml:"timeout"`
	SendBatchMaxSize int    `yaml:"send_batch_max_size"`
}

type MemoryLimiterConfig struct {
	CheckInterval        string `yaml:"check_interval"`
	LimitPercentage      int    `yaml:"limit_percentage"`
	SpikeLimitPercentage int    `yaml:"spike_limit_percentage"`
}

type ExtractK8sMetadataConfig struct {
	Metadata []string `yaml:"metadata"`
}

type PodAssociation struct {
	From string `yaml:"from"`
	Name string `yaml:"name,omitempty"`
}

type PodAssociations struct {
	Sources []PodAssociation `yaml:"sources"`
}

type K8sAttributesProcessorConfig struct {
	AuthType       string                   `yaml:"auth_type"`
	Passthrough    bool                     `yaml:"passthrough"`
	Extract        ExtractK8sMetadataConfig `yaml:"extract"`
	PodAssociation []PodAssociations        `yaml:"pod_association"`
}

type AttributeAction struct {
	Action string `yaml:"action"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
}

type ResourceProcessorConfig struct {
	Attributes []AttributeAction `yaml:"attributes"`
}

type ProcessorsConfig struct {
	Batch         BatchProcessorConfig         `yaml:"batch,omitempty"`
	MemoryLimiter MemoryLimiterConfig          `yaml:"memory_limiter,omitempty"`
	K8sAttributes K8sAttributesProcessorConfig `yaml:"k8sattributes,omitempty"`
	Resource      ResourceProcessorConfig      `yaml:"resource,omitempty"`
	Filter        FilterProcessorConfig        `yaml:"filter,omitempty"`
}

type FilterProcessorConfig struct {
	Traces TraceConfig `yaml:"traces,omitempty"`
}

type TraceConfig struct {
	Span []string `yaml:"span"`
}

type PipelineConfig struct {
	Receivers  []string `yaml:"receivers"`
	Processors []string `yaml:"processors"`
	Exporters  []string `yaml:"exporters"`
}

type PipelinesConfig struct {
	Traces PipelineConfig `yaml:"traces"`
}

type MetricsConfig struct {
	Address string `yaml:"address"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type TelemetryConfig struct {
	Metrics MetricsConfig `yaml:"metrics"`
	Logs    LoggingConfig `yaml:"logs"`
}

type OTLPServiceConfig struct {
	Pipelines  PipelinesConfig `yaml:"pipelines,omitempty"`
	Telemetry  TelemetryConfig `yaml:"telemetry,omitempty"`
	Extensions []string        `yaml:"extensions,omitempty"`
}

type ExtensionsConfig struct {
	HealthCheck map[string]any `yaml:"health_check"`
}

type OTELCollectorConfig struct {
	Receivers  ReceiverConfig    `yaml:"receivers"`
	Exporters  ExporterConfig    `yaml:"exporters"`
	Processors ProcessorsConfig  `yaml:"processors"`
	Extensions ExtensionsConfig  `yaml:"extensions"`
	Service    OTLPServiceConfig `yaml:"service"`
}

func makeReceiverConfig() ReceiverConfig {
	return ReceiverConfig{
		OpenCensus: map[string]any{},
		OTLP: map[string]any{
			"protocols": map[string]any{
				"http": map[string]any{},
				"grpc": map[string]any{},
			},
		},
	}
}

func getOutputType(output v1alpha1.TracePipelineOutput) string {
	if output.Otlp.Protocol == "http" {
		return "otlphttp"
	}
	return "otlp"
}

func makeHeaders(output v1alpha1.TracePipelineOutput) map[string]string {
	headers := make(map[string]string)
	if output.Otlp.Authentication != nil && output.Otlp.Authentication.Basic.IsDefined() {
		headers["Authorization"] = fmt.Sprintf("${%s}", basicAuthHeaderVariable)
	}
	for _, header := range output.Otlp.Headers {
		headers[header.Name] = fmt.Sprintf("${HEADER_%s}", envvar.MakeEnvVarCompliant(header.Name))
	}
	return headers
}

func makeExporterConfig(output v1alpha1.TracePipelineOutput, insecureOutput bool) ExporterConfig {
	outputType := getOutputType(output)
	headers := makeHeaders(output)
	otlpExporterConfig := OTLPExporterConfig{
		Endpoint: fmt.Sprintf("${%s}", otlpEndpointVariable),
		Headers:  headers,
		TLS: TLSConfig{
			Insecure: insecureOutput,
		},
		SendingQueue: SendingQueueConfig{
			Enabled:   true,
			QueueSize: 512,
		},
		RetryOnFailure: RetryOnFailureConfig{
			Enabled:         true,
			InitialInterval: "5s",
			MaxInterval:     "30s",
			MaxElapsedTime:  "300s",
		},
	}

	loggingExporter := LoggingExporterConfig{
		Verbosity: "basic",
	}

	if outputType == "otlphttp" {
		return ExporterConfig{
			OTLPHTTP: otlpExporterConfig,
			Logging:  loggingExporter,
		}
	}
	return ExporterConfig{
		OTLP:    otlpExporterConfig,
		Logging: loggingExporter,
	}
}

func makeProcessorsConfig() ProcessorsConfig {
	k8sAttributes := []string{
		"k8s.pod.name",
		"k8s.node.name",
		"k8s.namespace.name",
		"k8s.deployment.name",
		"k8s.statefulset.name",
		"k8s.daemonset.name",
		"k8s.cronjob.name",
		"k8s.job.name",
	}

	podAssociations := []PodAssociations{
		{
			Sources: []PodAssociation{
				{
					From: "resource_attribute",
					Name: "k8s.pod.ip",
				},
			},
		},
		{
			Sources: []PodAssociation{
				{
					From: "resource_attribute",
					Name: "k8s.pod.uid",
				},
			},
		},
		{
			Sources: []PodAssociation{
				{
					From: "connection",
				},
			},
		},
	}
	return ProcessorsConfig{
		Batch: BatchProcessorConfig{
			SendBatchSize:    512,
			Timeout:          "10s",
			SendBatchMaxSize: 512,
		},
		MemoryLimiter: MemoryLimiterConfig{
			CheckInterval:        "1s",
			LimitPercentage:      75,
			SpikeLimitPercentage: 10,
		},
		K8sAttributes: K8sAttributesProcessorConfig{
			AuthType:    "serviceAccount",
			Passthrough: false,
			Extract: ExtractK8sMetadataConfig{
				Metadata: k8sAttributes,
			},
			PodAssociation: podAssociations,
		},
		Resource: ResourceProcessorConfig{
			Attributes: []AttributeAction{
				{
					Action: "insert",
					Key:    "k8s.cluster.name",
					Value:  "${KUBERNETES_SERVICE_HOST}",
				},
			},
		},
		Filter: FilterProcessorConfig{
			Traces: TraceConfig{
				Span: makeSpanFilterConfig(),
			},
		},
	}
}

func makeSpanFilterConfig() []string {
	return []string{
		"(attributes[\"http.method\"] == \"POST\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (resource.attributes[\"service.name\"] == \"jaeger.kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Egress\") and (resource.attributes[\"service.name\"] == \"grafana.kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (resource.attributes[\"service.name\"] == \"jaeger.kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (resource.attributes[\"service.name\"] == \"grafana.kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (resource.attributes[\"service.name\"] == \"loki.kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (IsMatch(attributes[\"http.url\"], \".+/metrics\") == true) and (resource.attributes[\"k8s.namespace.name\"] == \"kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (IsMatch(attributes[\"http.url\"], \".+/healthz(/.*)?\") == true) and (resource.attributes[\"k8s.namespace.name\"] == \"kyma-system\")",
		"(attributes[\"http.method\"] == \"GET\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (attributes[\"user_agent\"] == \"vm_promscrape\")",
		"(attributes[\"http.method\"] == \"POST\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Egress\") and (IsMatch(attributes[\"http.url\"], \"http(s)?:\\\\/\\\\/telemetry-otlp-traces\\\\.kyma-system(\\\\..*)?:(4318|4317).*\") == true)",
		"(attributes[\"http.method\"] == \"POST\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Egress\") and (IsMatch(attributes[\"http.url\"], \"http(s)?:\\\\/\\\\/telemetry-trace-collector-internal\\\\.kyma-system(\\\\..*)?:(55678).*\") == true)",
		"(attributes[\"http.method\"] == \"POST\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Ingress\") and (resource.attributes[\"service.name\"] == \"loki.kyma-system\")",
		"(attributes[\"http.method\"] == \"POST\") and (attributes[\"component\"] == \"proxy\") and (attributes[\"OperationName\"] == \"Egress\") and (resource.attributes[\"service.name\"] == \"telemetry-fluent-bit.kyma-system\")",
	}
}

func makeServiceConfig(outputType string) OTLPServiceConfig {
	return OTLPServiceConfig{
		Pipelines: PipelinesConfig{
			Traces: PipelineConfig{
				Receivers:  []string{"opencensus", "otlp"},
				Processors: []string{"memory_limiter", "k8sattributes", "filter", "resource", "batch"},
				Exporters:  []string{outputType, "logging"},
			},
		},
		Telemetry: TelemetryConfig{
			Metrics: MetricsConfig{
				Address: "0.0.0.0:8888",
			},
			Logs: LoggingConfig{
				Level: "info",
			},
		},
		Extensions: []string{"health_check"},
	}
}

func makeExtensionConfig() ExtensionsConfig {
	return ExtensionsConfig{HealthCheck: map[string]any{}}
}

func makeOtelCollectorConfig(output v1alpha1.TracePipelineOutput, isInsecureOutput bool) OTELCollectorConfig {
	exporterConfig := makeExporterConfig(output, isInsecureOutput)
	outputType := getOutputType(output)
	processorsConfig := makeProcessorsConfig()
	receiverConfig := makeReceiverConfig()
	serviceConfig := makeServiceConfig(outputType)
	extensionConfig := makeExtensionConfig()

	return OTELCollectorConfig{
		Receivers:  receiverConfig,
		Exporters:  exporterConfig,
		Processors: processorsConfig,
		Extensions: extensionConfig,
		Service:    serviceConfig,
	}
}
