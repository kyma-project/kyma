# Alert Rules

## Overview

Kyma uses Prometheus alert rules for monitoring the health of its resources. This chart is the starting point for configuring alert rules.


### Creating Alert Rules in Kyma

Prometheus uses the a label selector **spec.ruleSelector** to identify those ConfigMap that holding Prometheus rule files.

```yaml
{{- if .Values.rulesSelector }}
  ruleSelector:
{{ toYaml .Values.rulesSelector | indent 4 }}
{{- else }}
  ruleSelector:
    matchLabels:
      role: alert-rules
      prometheus: {{ .Release.Name }}
{{- end }}
```

To define a new alert rule in Kyma, create a ConfigMap labelled with `role: alert-rules` as well as the name of the Prometheus object as `prometheus: {{ .Release.Name }}`.

Kyma provides the file [alert-rules-configmap.yaml](./templates/alert-rules-configmap.yaml) which serves as a reference to define Rules as configmaps.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: "Kyma"
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    prometheus: {{ .Release.Name }}
    release: {{ .Release.Name }}
    role: alert-rules
  name: {{ template "alert-rules.fullname" . }}
data:
{{- if .Values.prometheusRules }}
{{- $root := . }}
{{- range $key, $val := .Values.prometheusRules }}
  {{ $key }}: |-
{{ $val | indent 4}}
{{- end }}
{{ else }}
  alert.rules: |-
    {{- include "unhealthy-pods-rules.yaml.tpl" . | indent 4}}
{{ end }}
```
Under the **data. alert.rules** parameter, there is a configuration of the [unhealthy-pods-rules.yaml](templates/unhealthy-pods-rules.yaml) file, which creates a rule for alerting when a Pod is not running.

```yaml
{{ define "unhealthy-pods-rules.yaml.tpl" }}
groups:
- name: pod-not-running-rule
  rules:
  - alert: PodNotRunning
    expr: absent(kube_pod_container_status_running{namespace="default",pod="sample-metrics"})
    for: 15s
    labels:
      severity: critical
    annotations:
      description: "{{`{{$labels.namespace}}`}}/{{`{{$labels.pod}}`}} is not running"
      summary: "{{`{{$labels.pod}}`}} is not running"
{{ end }}
```
**A Quick explanation**
* ```alert:``` represents the name of the alert. Must be a valid metric name.
* ```expr:``` defines the PromQL expression to evaluate.
    - [kube_pod_container_status_running](https://github.com/kubernetes/kube-state-metrics/blob/master/Documentation/pod-metrics.md) is a [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) and in the expression above is evaluated if the pod, **pod="sample-metrics"** in the namespace, **namespace="default"** is running.
    - [Several functions](https://prometheus.io/docs/prometheus/latest/querying/functions/) are also provided by [Promethes](https://prometheus.io/docs/prometheus/latest/querying/basics/) to operate on data.
* ```for:``` Alerts are considered to be firing once they have been returned for this defined period of time.
* ```description:``` this annotation is used to enrich alert details.
* ```summary:``` this annotation is used to enrich alert details.

#### Generic resource metrics for pods

Resource metrics such as **cpu and memory** are also served by kube-state-metrics. The two metrics below are the Generic resource metrics recommended to be used in the future.

| Metric name| Metric type | Labels/tags |
| ---------- | ----------- | ----------- |
| kube_pod_container_resource_requests | Gauge | `resource`=&lt;resource-name&gt; <br> `unit`=&lt;resource-unit&gt; <br> `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |
| kube_pod_container_resource_limits | Gauge | `resource`=&lt;resource-name&gt; <br> `unit`=&lt;resource-unit&gt; <br> `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |

[Here](https://github.com/kubernetes/kube-state-metrics/blob/master/Documentation/pod-metrics.md) is the complete list of Pod Metrics


**Be aware that the metrics below will be removed in kube-state-metrics v2.0.0.**

- kube_pod_container_resource_requests_cpu_cores
- kube_pod_container_resource_limits_cpu_cores
- kube_pod_container_resource_requests_memory_bytes
- kube_pod_container_resource_limits_memory_bytes
- kube_pod_container_resource_requests_nvidia_gpu_devices
- kube_pod_container_resource_limits_nvidia_gpu_devices

### Monitoring Persistent Volume Claims
The [pvc-rules.yaml](templates/pvc-rules.yaml) configuration, located under the **data.alert.rules** parameter, specifes an alerting rule for the `PersistentVolumeClaim` (PVC). The Alertmanager triggers the rule when PVC in any of the system Namespaces such as `kyma-system`, `kyma-integration`, `heptio-ark`, `istio-system`, `kube-public` or `kube-system` exceeds  90%. In such a case, increase the capacity of PVCs.


### Configure Alertmanager

In Kyma all the configuration related to the Alertmanager is in the chart [alertmanager](../alertmanager/README.md)
