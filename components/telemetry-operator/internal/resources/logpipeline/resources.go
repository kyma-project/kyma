package logpipeline

import (
	"fmt"
	v1 "k8s.io/api/rbac/v1"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const checksumAnnotationKey = "checksum/logpipeline-config"

type DaemonSetConfig struct {
	FluentBitImage              string
	FluentBitConfigPrepperImage string
	ExporterImage               string
	PriorityClassName           string
	CPULimit                    resource.Quantity
	MemoryLimit                 resource.Quantity
	CPURequest                  resource.Quantity
	MemoryRequest               resource.Quantity
}

func MakeServiceAccount(name types.NamespacedName) *corev1.ServiceAccount {
	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
	}
	return &serviceAccount
}

func MakeClusterRoleBinding(name types.NamespacedName) *v1.ClusterRoleBinding {
	clusterRoleBinding := v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Subjects: []v1.Subject{{Name: "telemetry-fluent-bit", Namespace: "kyma-system", Kind: "ServiceAccount"}},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     name.Name,
		},
	}
	return &clusterRoleBinding
}

func MakeClusterRole(name types.NamespacedName) *v1.ClusterRole {
	clusterRole := v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Rules: []v1.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{""}, Resources: []string{"namespaces", "pods"}}},
	}
	return &clusterRole
}

func MakeDaemonSet(name types.NamespacedName, checksum string, dsConfig DaemonSetConfig) *appsv1.DaemonSet {
	resourcesFluentBit := corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    dsConfig.CPURequest,
			corev1.ResourceMemory: dsConfig.MemoryRequest,
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    dsConfig.CPULimit,
			corev1.ResourceMemory: dsConfig.MemoryLimit,
		},
	}

	resourcesExporter := corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceMemory: resource.MustParse("5Mi"),
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("50Mi"),
		},
	}

	annotations := make(map[string]string)
	annotations[checksumAnnotationKey] = checksum
	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Labels:    labels(),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels(),
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: name.Name,
					PriorityClassName:  dsConfig.PriorityClassName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot:   pointer.Bool(false),
						SeccompProfile: &corev1.SeccompProfile{Type: "RuntimeDefault"},
					},
					Containers: []corev1.Container{
						{
							Name: "fluent-bit",
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Capabilities: &corev1.Capabilities{
									Add:  []corev1.Capability{"FOWNER"},
									Drop: []corev1.Capability{"ALL"},
								},
								Privileged:             pointer.Bool(false),
								ReadOnlyRootFilesystem: pointer.Bool(true),
							},
							Image:           dsConfig.FluentBitImage,
							ImagePullPolicy: "IfNotPresent",
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-env", name.Name)},
										Optional:             pointer.Bool(true),
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 2020,
									Protocol:      "TCP",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromString("http"),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/v1/health",
										Port: intstr.FromString("http"),
									},
								},
							},
							Resources: resourcesFluentBit,
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/fluent-bit/etc", Name: "shared-fluent-bit-config"},
								{MountPath: "/fluent-bit/etc/fluent-bit.conf", Name: "config", SubPath: "fluent-bit.conf"},
								{MountPath: "/fluent-bit/etc/dynamic/", Name: "dynamic-config"},
								{MountPath: "/fluent-bit/etc/dynamic-parsers/", Name: "dynamic-parsers-config"},
								{MountPath: "/fluent-bit/etc/custom_parsers.conf", Name: "config", SubPath: "custom_parsers.conf"},
								{MountPath: "/fluent-bit/etc/loki-labelmap.json", Name: "config", SubPath: "loki-labelmap.json"},
								{MountPath: "/fluent-bit/scripts/filter-script.lua", Name: "luascripts", SubPath: "filter-script.lua"},
								{MountPath: "/var/log", Name: "varlog"},
								{MountPath: "/var/lib/docker/containers", Name: "varlibdockercontainers", ReadOnly: true},
								{MountPath: "/data", Name: "varfluentbit"},
								{MountPath: "/files", Name: "dynamic-files"},
							},
						},
						{
							Name:      "exporter",
							Image:     dsConfig.ExporterImage,
							Resources: resourcesExporter,
							Args: []string{
								"--storage-path=/data/flb-storage/",
								"--metric-name=telemetry_fsbuffer_usage_bytes",
							},
							WorkingDir: "",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http-metrics",
									ContainerPort: 2021,
									Protocol:      "TCP",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Privileged:               pointer.Bool(false),
								ReadOnlyRootFilesystem:   pointer.Bool(true),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "varfluentbit", MountPath: "/data"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: name.Name},
								},
							},
						},
						{
							Name: "luascripts",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-luascripts", name.Name)},
								},
							},
						},
						{
							Name: "varlog",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{Path: "/var/log"},
							},
						},
						{
							Name: "varlibdockercontainers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{Path: "/var/lib/docker/containers"},
							},
						},
						{
							Name: "shared-fluent-bit-config",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "dynamic-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-sections", name.Name)},
									Optional:             pointer.Bool(true),
								},
							},
						},
						{
							Name: "dynamic-parsers-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-parsers", name.Name)},
									Optional:             pointer.Bool(true),
								},
							},
						},
						{
							Name: "dynamic-files",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-files", name.Name)},
									Optional:             pointer.Bool(true),
								},
							},
						},
						{
							Name: "varfluentbit",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{Path: fmt.Sprintf("/var/%s", name.Name)},
							},
						},
					},
				},
			},
		},
	}
}

func MakeMetricsService(name types.NamespacedName) *corev1.Service {
	metricsPort := 2020
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-metrics", name.Name),
			Namespace: name.Namespace,
			Labels:    labels(),
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   strconv.Itoa(metricsPort),
				"prometheus.io/scheme": "http",
				"prometheus.io/path":   "/api/v1/metrics/prometheus",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       int32(metricsPort),
					TargetPort: intstr.FromString("http"),
				},
			},
			Selector: labels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func MakeExporterMetricsService(name types.NamespacedName) *corev1.Service {
	metricsPort := 2021
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-exporter-metrics", name.Name),
			Namespace: name.Namespace,
			Labels:    labels(),
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   strconv.Itoa(metricsPort),
				"prometheus.io/scheme": "http",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http-metrics",
					Protocol:   "TCP",
					Port:       int32(metricsPort),
					TargetPort: intstr.FromString("http-metrics"),
				},
			},
			Selector: labels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func MakeConfigMap(name types.NamespacedName) *corev1.ConfigMap {
	parserConfig := `
[PARSER]
    Name docker_no_time
    Format json
    Time_Keep Off
    Time_Key time
    Time_Format %Y-%m-%dT%H:%M:%S.%L
`

	fluentBitConfig := `
[SERVICE]
    Daemon Off
    Flush 1
    Log_Level warn
    Parsers_File custom_parsers.conf
    Parsers_File dynamic-parsers/parsers.conf
    HTTP_Server On
    HTTP_Listen 0.0.0.0
    HTTP_Port 2020
    storage.path /data/flb-storage/
    storage.metrics on

[INPUT]
    Name tail
    Alias tele-tail
    Path /var/log/containers/*.log
    multiline.parser docker, cri, go, python, java
    Tag tele.*
    Mem_Buf_Limit 5MB
    Skip_Long_Lines On
    Refresh_Interval 10
    DB /data/flb_kube.db
    storage.type  filesystem
    Read_from_Head True

[FILTER]
    Name kubernetes
    Match tele.*
    Merge_Log On
    K8S-Logging.Parser On
    K8S-Logging.Exclude On
    Buffer_Size 1MB

@INCLUDE dynamic/*.conf
`
	lokiLabelmap := `
  {
    "kubernetes": {
      "container_name": "container",
      "host": "node",
      "labels": {
        "app": "app",
        "app.kubernetes.io/component": "component",
        "app.kubernetes.io/name": "app",
        "serverless.kyma-project.io/function-name": "function"
      },
      "namespace_name": "namespace",
      "pod_name": "pod"
    },
    "stream": "stream"
  }
`

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Labels:    labels(),
		},
		Data: map[string]string{
			"custom_parsers.conf": parserConfig,
			"fluent-bit.conf":     fluentBitConfig,
			"loki-labelmap.json":  lokiLabelmap,
		},
	}
}

func MakeDynamicParserConfigmap(name types.NamespacedName) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-parsers", name.Name),
			Namespace: name.Namespace,
			Labels:    labels(),
		},
		Data: map[string]string{"parsers.conf": ""},
	}
}

func MakeLuaConfigMap(name types.NamespacedName) *corev1.ConfigMap {
	luaFilter := `
function kubernetes_map_keys(tag, timestamp, record)
  if record.kubernetes == nil then
    return 0
  end
  map_keys(record.kubernetes.annotations)
  map_keys(record.kubernetes.labels)
  return 1, timestamp, record
end
function map_keys(table)
  if table == nil then
    return
  end
  local new_table = {}
  local changed_keys = {}
  for key, val in pairs(table) do
    local mapped_key = string.gsub(key, "[%/%.]", "_")
    if mapped_key ~= key then
      new_table[mapped_key] = val
      changed_keys[key] = true
    end
  end
  for key in pairs(changed_keys) do
    table[key] = nil
  end
  for key, val in pairs(new_table) do
    table[key] = val
  end
end
`

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-luascripts", name.Name),
			Namespace: name.Namespace,
			Labels:    labels(),
		},
		Data: map[string]string{"filter-script.lua": luaFilter},
	}
}

func labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "fluent-bit",
		"app.kubernetes.io/instance": "telemetry",
	}
}
