package tracepipeline

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

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

func makeConfigMap(config Config, collectorConfig OTELCollectorConfig) *corev1.ConfigMap {
	bytes, _ := yaml.Marshal(collectorConfig)
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
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "8888",
			},
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
