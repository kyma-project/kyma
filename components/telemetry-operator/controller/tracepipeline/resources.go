package tracepipeline

import (
	"fmt"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	basicAuthHeaderVariable = "BASIC_AUTH_HEADER"
	otlpEndpointVariable    = "OTLP_ENDPOINT"
	configHashAnnotationKey = "checksum/config"
)

var (
	collectorResources = corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("150m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("1000m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}
	configMapKey          = "relay.conf"
	defaultPodAnnotations = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	replicas = int32(1)
)

func getLabels(config Config) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": config.ResourceName,
	}
}

func makeConfigMap(config Config, output v1alpha1.TracePipelineOutput) *corev1.ConfigMap {
	exporterConfig := makeExporterConfig(output)
	outputType := getOutputType(output)
	processorsConfig := makeProcessorsConfig()
	conf := confmap.NewFromStringMap(map[string]any{
		"receivers": map[string]any{
			"opencensus": map[string]any{},
			"otlp": map[string]any{
				"protocols": map[string]any{
					"http": map[string]any{},
					"grpc": map[string]any{},
				},
			},
		},
		"exporters":  exporterConfig,
		"processors": processorsConfig,
		"service": map[string]any{
			"pipelines": map[string]any{
				"traces": map[string]any{
					"receivers":  []any{"opencensus", "otlp"},
					"processors": []any{"memory_limiter", "batch"},
					"exporters":  []any{outputType},
				},
			},
			"telemetry": map[string]any{
				"metrics": map[string]any{
					"address": "0.0.0.0:8888",
				},
			},
		},
	})

	bytes, _ := yaml.Marshal(conf.ToStringMap())
	confYAML := string(bytes)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ResourceName,
			Namespace: config.CollectorNamespace,
			Labels:    getLabels(config),
		},
		Data: map[string]string{
			configMapKey: confYAML,
		},
	}
}

func makeSecret(config Config, secretData map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ResourceName,
			Namespace: config.CollectorNamespace,
			Labels:    getLabels(config),
		},
		Data: secretData,
	}
}

func getOutputType(output v1alpha1.TracePipelineOutput) string {
	if output.Otlp.Protocol == "http" {
		return "otlphttp"
	}
	return "otlp"
}

func makeExporterConfig(output v1alpha1.TracePipelineOutput) map[string]any {
	outputType := getOutputType(output)
	var headers map[string]any
	if output.Otlp.Authentication != nil && output.Otlp.Authentication.Basic.IsDefined() {
		headers = map[string]any{
			"Authorization": fmt.Sprintf("${%s}", basicAuthHeaderVariable),
		}
	}
	return map[string]any{
		outputType: map[string]any{
			"endpoint": fmt.Sprintf("${%s}", otlpEndpointVariable),
			"headers":  headers,
			"sending_queue": map[string]any{
				"enabled":    true,
				"queue_size": 512,
			},
			"retry_on_failure": map[string]any{
				"enabled":          true,
				"initial_interval": "5s",
				"max_interval":     "30s",
				"max_elapsed_time": "300s",
			},
		},
	}
}

func makeProcessorsConfig() map[string]any {
	return map[string]any{
		"batch": map[string]any{
			"send_batch_size":     512,
			"timeout":             "10s",
			"send_batch_max_size": 512,
		},
		"memory_limiter": map[string]any{
			"check_interval":         "1s",
			"limit_percentage":       75,
			"spike_limit_percentage": 10,
		},
	}
}

func makeDeployment(config Config, configHash string) *appsv1.Deployment {
	labels := getLabels(config)
	optional := true
	annotations := makePodAnnotations(configHash)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ResourceName,
			Namespace: config.CollectorNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  config.ResourceName,
							Image: config.CollectorImage,
							Args:  []string{"--config=/conf/" + configMapKey},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: config.ResourceName,
										},
										Optional: &optional,
									},
								},
							},
							Resources:    collectorResources,
							VolumeMounts: []corev1.VolumeMount{{Name: "config", MountPath: "/conf"}},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: config.ResourceName,
									},
									Items: []corev1.KeyToPath{{Key: configMapKey, Path: configMapKey}},
								},
							},
						},
					},
				},
			},
		},
	}
}

func makePodAnnotations(configHash string) map[string]string {
	annotations := map[string]string{
		configHashAnnotationKey: configHash,
	}
	for k, v := range defaultPodAnnotations {
		annotations[k] = v
	}
	return annotations
}

func makeCollectorService(config Config) *corev1.Service {
	labels := getLabels(config)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ResourceName,
			Namespace: config.CollectorNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "grpc-otlp",
					Protocol:   corev1.ProtocolTCP,
					Port:       4317,
					TargetPort: intstr.FromInt(4317),
				},
				{
					Name:       "http-otlp",
					Protocol:   corev1.ProtocolTCP,
					Port:       4318,
					TargetPort: intstr.FromInt(4318),
				},
				{
					Name:       "http-opencensus",
					Protocol:   corev1.ProtocolTCP,
					Port:       55678,
					TargetPort: intstr.FromInt(55678),
				},
			},
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func makeMetricsService(config Config) *corev1.Service {
	labels := getLabels(config)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ResourceName + "-metrics",
			Namespace: config.CollectorNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http-metrics",
					Protocol:   corev1.ProtocolTCP,
					Port:       8888,
					TargetPort: intstr.FromInt(8888),
				},
			},
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func makeServiceMonitor(config Config) *monitoringv1.ServiceMonitor {
	labels := getLabels(config)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ResourceName,
			Namespace: config.CollectorNamespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port: "http-metrics",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					config.CollectorNamespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
}
