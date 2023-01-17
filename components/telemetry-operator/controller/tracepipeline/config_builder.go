package tracepipeline

import (
	"fmt"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type TlsConfig struct {
	Insecure bool `yaml:"insecure"`
}

type OtlpExporterConfig struct {
	Endpoint       string         `yaml:"endpoint,omitempty"`
	Headers        map[string]any `yaml:"headers,omitempty"`
	Tls            TlsConfig      `yaml:"tls,omitempty"`
	SendingQueue   map[string]any `yaml:"sending_queue,omitempty"`
	RetryOnFailure map[string]any `yaml:"retry_on_failure,omitempty"`
}

type ExporterConfig struct {
	Otlp     OtlpExporterConfig `yaml:"otlp,omitempty"`
	OtlpHttp OtlpExporterConfig `yaml:"otlphttp,omitempty"`
}

type ReceiverConfig struct {
	OpenCensus map[string]any `yaml:"opencensus"`
	Otlp       map[string]any `yaml:"otlp"`
}

type ProcessorsConfig struct {
	Batch         map[string]any `yaml:"batch,omitempty"`
	MemoryLimiter map[string]any `yaml:"memory_limiter,omitempty"`
	K8sAttributes map[string]any `yaml:"k8sattributes,omitempty"`
	Resource      map[string]any `yaml:"resource,omitempty"`
}

type PipelineConfig struct {
	Receivers  []string `yaml:"receivers"`
	Processors []string `yaml:"processors"`
	Exporters  []string `yaml:"exporters"`
}

type PipelinesConfig struct {
	Traces PipelineConfig `yaml:"traces"`
}

type OtlpServiceConfig struct {
	Pipelines  PipelinesConfig `yaml:"pipelines,omitempty"`
	Telemetry  map[string]any  `yaml:"telemetry,omitempty"`
	Extensions []string        `yaml:"extensions,omitempty"`
}

type ExtensionsConfig struct {
	HealthCheck map[string]any `yaml:"health_check"`
}

type OtelCollectorConfig struct {
	Receivers  ReceiverConfig    `yaml:"receivers"`
	Exporters  ExporterConfig    `yaml:"exporters"`
	Processors ProcessorsConfig  `yaml:"processors"`
	Extensions ExtensionsConfig  `yaml:"extensions"`
	Service    OtlpServiceConfig `yaml:"service"`
}

func makeReceiverConfig() ReceiverConfig {
	return ReceiverConfig{
		OpenCensus: map[string]any{},
		Otlp: map[string]any{
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

func makeExporterConfig(output v1alpha1.TracePipelineOutput, insecureOutput bool) ExporterConfig {
	outputType := getOutputType(output)
	var headers map[string]any
	if output.Otlp.Authentication != nil && output.Otlp.Authentication.Basic.IsDefined() {
		headers = map[string]any{
			"Authorization": fmt.Sprintf("${%s}", basicAuthHeaderVariable),
		}
	}
	otlpExporterConfig := OtlpExporterConfig{
		Endpoint: fmt.Sprintf("${%s}", otlpEndpointVariable),
		Headers:  headers,
		Tls: TlsConfig{
			Insecure: insecureOutput,
		},
		SendingQueue: map[string]any{
			"enabled":    true,
			"queue_size": 512,
		},
		RetryOnFailure: map[string]any{
			"enabled":          true,
			"initial_interval": "5s",
			"max_interval":     "30s",
			"max_elapsed_time": "300s",
		},
	}

	if outputType == "otlphttp" {
		return ExporterConfig{
			OtlpHttp: otlpExporterConfig,
		}
	}
	return ExporterConfig{
		Otlp: otlpExporterConfig,
	}
}

func makeProcessorsConfig() ProcessorsConfig {
	k8sAttributes := []any{
		"k8s.pod.name",
		"k8s.node.name",
		"k8s.namespace.name",
		"k8s.deployment.name",
		"k8s.statefulset.name",
		"k8s.daemonset.name",
		"k8s.cronjob.name",
		"k8s.job.name",
	}

	podAssociations := []map[string]any{
		{
			"sources": []map[string]any{{
				"from": "resource_attribute",
				"name": "k8s.pod.ip",
			},
			},
		},
		{
			"sources": []map[string]any{{
				"from": "resource_attribute",
				"name": "k8s.pod.uid",
			},
			},
		},
		{
			"sources": []map[string]any{{
				"from": "connection",
			},
			},
		},
	}
	return ProcessorsConfig{
		Batch: map[string]any{
			"send_batch_size":     512,
			"timeout":             "10s",
			"send_batch_max_size": 512,
		},
		MemoryLimiter: map[string]any{
			"check_interval":         "1s",
			"limit_percentage":       75,
			"spike_limit_percentage": 10,
		},
		K8sAttributes: map[string]any{
			"auth_type":   "serviceAccount",
			"passthrough": "false",
			"extract": map[string]any{
				"metadata": k8sAttributes,
			},
			"pod_association": podAssociations,
		},
		Resource: map[string]any{
			"attributes": []map[string]any{
				{
					"action": "insert",
					"key":    "k8s.cluster.name",
					"value":  "${KUBERNETES_SERVICE_HOST}",
				},
			},
		},
	}
}

func makeServiceConfig(outputType string) OtlpServiceConfig {
	return OtlpServiceConfig{
		Pipelines: PipelinesConfig{
			Traces: PipelineConfig{
				Receivers:  []string{"opencensus", "otlp"},
				Processors: []string{"memory_limiter", "k8sattributes", "resource", "batch"},
				Exporters:  []string{outputType},
			},
		},
		Telemetry: map[string]any{
			"metrics": map[string]any{
				"address": "0.0.0.0:8888",
			},
		},
		Extensions: []string{"health_check"},
	}
}

func makeExtensionConfig() ExtensionsConfig {
	return ExtensionsConfig{HealthCheck: map[string]any{}}
}

func makeOtelCollectorConfig(output v1alpha1.TracePipelineOutput, isInsecureOutput bool) OtelCollectorConfig {
	exporterConfig := makeExporterConfig(output, isInsecureOutput)
	outputType := getOutputType(output)
	processorsConfig := makeProcessorsConfig()
	receiverConfig := makeReceiverConfig()
	serviceConfig := makeServiceConfig(outputType)
	extensionConfig := makeExtensionConfig()

	return OtelCollectorConfig{
		Receivers:  receiverConfig,
		Exporters:  exporterConfig,
		Processors: processorsConfig,
		Extensions: extensionConfig,
		Service:    serviceConfig,
	}
}
