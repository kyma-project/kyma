package tracepipeline

import (
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

var (
	podAnnotations = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	replicas           = int32(1)
	collectorResources = corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("10m"),
			corev1.ResourceMemory: resource.MustParse("64Mi"),
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
)

func getLabels(config Config) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": config.CollectorDeploymentName,
	}
}

func makeConfigMap(config Config, output v1alpha1.TracePipelineOutput) *corev1.ConfigMap {
	exporterConfig := makeExporterConfig(output)
	outputType := getOutputType(output)
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
		"exporters": exporterConfig,
		"service": map[string]any{
			"pipelines": map[string]any{
				"traces": map[string]any{
					"receivers": []any{"opencensus", "otlp"},
					"exporters": []any{outputType},
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
			Name:      config.CollectorConfigMapName,
			Namespace: config.CollectorNamespace,
			Labels:    getLabels(config),
		},
		Data: map[string]string{
			config.ConfigMapKey: confYAML,
		},
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
	return map[string]any{
		outputType: map[string]any{
			"endpoint": output.Otlp.Endpoint.Value,
		},
	}
}

func makeDeployment(config Config) *appsv1.Deployment {
	labels := getLabels(config)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CollectorDeploymentName,
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
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    config.CollectorDeploymentName,
							Image:   config.CollectorImage,
							Command: []string{"/otelcol-contrib", "--config=/conf/" + config.ConfigMapKey},
							Env: []corev1.EnvVar{
								{
									Name: "MY_POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "status.podIP",
										},
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
										Name: config.CollectorConfigMapName,
									},
									Items: []corev1.KeyToPath{{Key: config.ConfigMapKey, Path: config.ConfigMapKey}},
								},
							},
						},
					},
				},
			},
		},
	}
}

func makeCollectorService(config Config) *corev1.Service {
	labels := getLabels(config)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CollectorDeploymentName,
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
			Name:      config.CollectorDeploymentName + "-metrics",
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
			Name:      config.CollectorDeploymentName,
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
