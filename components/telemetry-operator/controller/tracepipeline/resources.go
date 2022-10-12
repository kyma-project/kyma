package tracepipeline

import (
	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CollectorDeploymentName,
			Namespace: config.CollectorNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &config.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: config.PodSelectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      config.PodSelectorLabels,
					Annotations: config.PodAnnotations,
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
							Resources:    config.CollectorResources,
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

func makeService(config Config) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CollectorDeploymentName,
			Namespace: config.CollectorNamespace,
			Labels:    config.PodSelectorLabels,
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
				{
					Name:       "http-metrics",
					Protocol:   corev1.ProtocolTCP,
					Port:       8888,
					TargetPort: intstr.FromInt(8888),
				},
			},
			Selector: config.PodSelectorLabels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func makeServiceMonitor(config Config) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CollectorDeploymentName,
			Namespace: config.CollectorNamespace,
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
				MatchLabels: config.PodSelectorLabels,
			},
		},
	}
}
