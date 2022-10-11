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
	name              = "opentelemetry-collector"
	configMapName     = "opentelemetry-collector-config"
	configMapKey      = "relay.conf"
	systemNamespace   = "kyma-system"
	podSelectorLabels = map[string]string{
		"app.kubernetes.io/name": name,
	}
	podAnnotations = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
)

func makeConfigMap(output v1alpha1.TracePipelineOutput) *corev1.ConfigMap {
	conf := confmap.NewFromStringMap(map[string]any{
		"receivers": map[string]any{
			"opencensus": map[string]any{},
			"otlp": map[string]any{
				"protocols": map[string]any{
					"http": map[string]any{},
				},
			},
		},
		"exporters": map[string]any{
			"otlphttp": map[string]any{
				"endpoint": output.Otlp.URL.Value,
			},
		},
		"service": map[string]any{
			"pipelines": map[string]any{
				"traces": map[string]any{
					"receivers": []any{"opencensus", "otlp"},
					"exporters": []any{"otlphttp"},
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
			Name:      configMapName,
			Namespace: systemNamespace,
		},
		Data: map[string]string{
			configMapKey: confYAML,
		},
	}
}

func makeDeployment() *appsv1.Deployment {
	var replicaCount int32 = 1
	image := "otel/opentelemetry-collector-contrib:0.60.0"

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: systemNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: podSelectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podSelectorLabels,
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    name,
							Image:   image,
							Command: []string{"/otelcol-contrib", "--config=/conf/relay.yaml"},
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
							Resources: corev1.ResourceRequirements{
								Limits: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    resource.MustParse("256m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{{Name: "config", MountPath: "/conf"}},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
									Items: []corev1.KeyToPath{{Key: configMapKey, Path: "relay.yaml"}},
								},
							},
						},
					},
				},
			},
		},
	}
}

func makeService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: systemNamespace,
			Labels:    podSelectorLabels,
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
			Selector: podSelectorLabels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func makeServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: systemNamespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port: "http-metrics",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					systemNamespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: podSelectorLabels,
			},
		},
	}
}
