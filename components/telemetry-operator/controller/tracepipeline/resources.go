package tracepipeline

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	basicAuthHeaderVariable = "BASIC_AUTH_HEADER"
	otlpEndpointVariable    = "OTLP_ENDPOINT"
	configHashAnnotationKey = "checksum/config"
	collectorUser           = 10001
	collectorContainerName  = "collector"
)

var (
	configMapKey          = "relay.conf"
	defaultPodAnnotations = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	replicas = int32(1)
)

func makeDefaultLabels(config Config) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": config.BaseName,
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
		"extensions": map[string]any{
			"health_check": map[string]any{},
		},
		"service": map[string]any{
			"pipelines": map[string]any{
				"traces": map[string]any{
					"receivers":  []any{"opencensus", "otlp"},
					"processors": []any{"memory_limiter", "k8sattributes", "resource", "batch"},
					"exporters":  []any{outputType},
				},
			},
			"telemetry": map[string]any{
				"metrics": map[string]any{
					"address": "0.0.0.0:8888",
				},
			},
			"extensions": []string{"health_check"},
		},
	})

	bytes, _ := yaml.Marshal(conf.ToStringMap())
	confYAML := string(bytes)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.BaseName,
			Namespace: config.Namespace,
			Labels:    makeDefaultLabels(config),
		},
		Data: map[string]string{
			configMapKey: confYAML,
		},
	}
}

func makeSecret(config Config, secretData map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.BaseName,
			Namespace: config.Namespace,
			Labels:    makeDefaultLabels(config),
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
	isInsecure := len(strings.TrimSpace(output.Otlp.Endpoint.Value)) > 0 && strings.HasPrefix(output.Otlp.Endpoint.Value, "http://")
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
			"tls": map[string]any{
				"insecure": isInsecure,
			},
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
		"k8sattributes": map[string]any{
			"auth_type":   "serviceAccount",
			"passthrough": "false",
			"extract": map[string]any{
				"metadata": k8sAttributes,
			},
			"pod_association": podAssociations,
		},
		"resource": map[string]any{
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

func makeDeployment(config Config, configHash string) *appsv1.Deployment {
	labels := makeDefaultLabels(config)
	optional := true
	annotations := makePodAnnotations(configHash)
	resources := makeResourceRequirements(config)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.BaseName,
			Namespace: config.Namespace,
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
							Name:  collectorContainerName,
							Image: config.Deployment.Image,
							Args:  []string{"--config=/conf/" + configMapKey},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: config.BaseName,
										},
										Optional: &optional,
									},
								},
							},
							Resources: resources,
							SecurityContext: &corev1.SecurityContext{
								Privileged:               pointer.Bool(false),
								RunAsUser:                pointer.Int64(collectorUser),
								RunAsNonRoot:             pointer.Bool(true),
								ReadOnlyRootFilesystem:   pointer.Bool(true),
								AllowPrivilegeEscalation: pointer.Bool(false),
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
							VolumeMounts: []corev1.VolumeMount{{Name: "config", MountPath: "/conf"}},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.IntOrString{IntVal: 13133}},
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.IntOrString{IntVal: 13133}},
								},
							},
						},
					},
					ServiceAccountName: config.BaseName,
					PriorityClassName:  config.Deployment.PriorityClassName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    pointer.Int64(collectorUser),
						RunAsNonRoot: pointer.Bool(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: config.BaseName,
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

func makeResourceRequirements(config Config) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    config.Deployment.CPURequest,
			corev1.ResourceMemory: config.Deployment.MemoryRequest,
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    config.Deployment.CPULimit,
			corev1.ResourceMemory: config.Deployment.MemoryLimit,
		},
	}
}

func makeOTLPService(config Config) *corev1.Service {
	labels := makeDefaultLabels(config)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Service.OTLPServiceName,
			Namespace: config.Namespace,
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
			},
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func makeMetricsService(config Config) *corev1.Service {
	labels := makeDefaultLabels(config)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.BaseName + "-metrics",
			Namespace: config.Namespace,
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

func makeOpenCensusService(config Config) *corev1.Service {
	labels := makeDefaultLabels(config)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.BaseName + "-internal",
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
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

func makeServiceMonitor(config Config) *monitoringv1.ServiceMonitor {
	labels := makeDefaultLabels(config)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.BaseName,
			Namespace: config.Namespace,
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
					config.Namespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
}
