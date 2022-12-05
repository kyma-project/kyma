package logpipeline

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type resourceConfig struct {
	Name      types.NamespacedName
	Component string
}

func makeDaemonSet(config resourceConfig) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name(config),
			Labels: labels(config),
		},
		Spec:   appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(config),
			},
			Template:             corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:                     labels(config),
					Annotations:                map[string]string{
						"checksum/config" : "5c928f5bb13aabb32f086b320b200c0df16c9ca8a553a856bad3993cd3582143",
						"checksum/luascripts": "d538085e16b852526ec9e36ce06000a0bb6907347d083db72c15ce0952bda559",
					},
				},
				Spec:       corev1.PodSpec{
					ServiceAccountName:            name(config),
					PriorityClassName:             "kyma-system-priority",
					SecurityContext:               &corev1.PodSecurityContext{
						RunAsNonRoot:        nil, //TODO
						SeccompProfile:      nil, //TODO
					},
					HostNetwork:                   false,
					InitContainers:                nil, //TODO
					Containers:                    nil, //TODO
					Volumes:                       nil, //TODO
				},
			},
		},
	}
}

func makeService(config resourceConfig) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name(config),
			Namespace: config.Name.Namespace,
			Labels: labels(config),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       2020,
					TargetPort: intstr.IntOrString{StrVal: "http"},
				},
				{
					Name:       "http-metrics",
					Port:       2021,
					TargetPort: intstr.IntOrString{StrVal: "http-metrics"},
				},
			},
			Selector: labels(config),
		},
	}
}

func makeConfigMap(config resourceConfig) *corev1.ConfigMap {
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
    Parsers_File parsers.conf
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

[INPUT]
    Name tail
    Path /null.log
    Tag null.*
    Alias null-tail

[FILTER]
    Name kubernetes
    Match tele.*
    Merge_Log On
    K8S-Logging.Parser On
    K8S-Logging.Exclude On
    Buffer_Size 1MB

[OUTPUT]
    Name null
    Match null.*
    Alias null-null

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
			Name:      name(config),
			Namespace: config.Name.Namespace,
			Labels: labels(config),
		},
		Data: map[string]string{
			"custom_parsers.conf": parserConfig,
			"fluent-bit.conf":     fluentBitConfig,
			"loki-labelmap.json":  lokiLabelmap,
		},
	}
}

func makeLuaConfigMap(config resourceConfig) *corev1.ConfigMap {
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
    local mapped_key = string.gsub(key, \"[%/%.]\", \"_\")
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
			Name:      fmt.Sprintf("%s-%s-luascripts", config.Component, config.Name.Name),
			Namespace: config.Name.Namespace,
			Labels: labels(config),
		},
		Data: map[string]string{"filter-script.lua": luaFilter},
	}
}

func labels(config resourceConfig) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     config.Name.Name,
		"app.kubernetes.io/instance": config.Component,
	}
}

func name(config resourceConfig) string {
	return fmt.Sprintf("%s-%s", config.Component, config.Name.Name)
}
