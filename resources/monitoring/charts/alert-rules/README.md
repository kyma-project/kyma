# Alert Rules

## Overview

Kyma uses Prometheus alert rules to monitor the health of its resources. Use this chart to configure alert rules.

## Alert rules

You can define the following alert rules:

- Pod is not running

    The Alertmanager sends out alerts when one of the Pods is not running in `kyma-system`, `kyma-integration`, `istio-system`, `kube-public`, or `kube-system` Namespaces.

- Monitor Persistent Volume Claims (PVC)

    The Alertmanager triggers the rule when PVC exceeds 90% for the following system Namespaces: `kyma-system`, `kyma-integration`, `heptio-ark`, `istio-system`, `kube-public`, or `kube-system`. To avoid this, increase the capacity of PVC.

-  Monitor CPU Usage

    The Alertmanager triggers the rule when CPU usage exceeds 90% for Pods in the `kyma-system` Namespace. For the alert rule to activate, make sure to add the `alertcpu: "yes"` label to Pods.

- Monitor Memory usage

    The Alertmanager triggers the rule when Memory usage exceeds 90% for Pods in the `kyma-system` Namespace. For the alert rule to activate, make sure to add the `alertmem: "yes"` to Pods.

## Create alert rules

Prometheus uses the  **spec.ruleSelector** label selector to identify ConfigMaps which include Prometheus rule files.

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
Follow the steps to create an alert rule:

1. Use the [ConfigMap template](./templates/alert-rules-configmap.yaml) which contains the sample configuration for an alert rule.


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
    {{- include "kyma-rules.yaml.tpl" . | indent 4}}
{{ end }}
```

2. Label the ConfigMap with `role: alert-rules`.
3. Add the name of a Prometheus object in `prometheus: {{ .Release.Name }}`.
3. Configure the **data. alert.rules** parameter in the[kyma-rules.yaml](templates/kyma-rules.yaml) file. 


The example shows a sample configuration for an alert rule. The rule activates the alarm when a Pod is not running.

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
The rule definition includes the following parameters:

- **alert:** is the valid metric name name of the alert.
- **expr:** defines the PromQL expression to evaluate. The expression uses the `kube_pod_container_status_running` Pod metric. In the example, it is used to check if the `sample-metrics` Pod is running in the `default` Namespace. Prometheus provides additional [functions](https://prometheus.io/docs/prometheus/latest/querying/functions/) you can use to operate on data.
* **for:**  is a period of time within which alerts are returned.
* **description:** is an annotation used to enrich alert details.
* **summary:** is an annotation used to enrich alert details.

#### Generic resource metrics for Pods

The **cpu and memory** metrics are generic resource metrics.

| Metric name| Metric type | Labels/tags |
| ---------- | ----------- | ----------- |
| kube_pod_container_resource_requests | Gauge | `resource`=&lt;resource-name&gt; <br> `unit`=&lt;resource-unit&gt; <br> `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |
| kube_pod_container_resource_limits | Gauge | `resource`=&lt;resource-name&gt; <br> `unit`=&lt;resource-unit&gt; <br> `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |

[Here](https://github.com/kubernetes/kube-state-metrics/blob/master/Documentation/pod-metrics.md) is the complete list of Pod Metrics.


`kube-state-metrics` v2.0.0 does not include the following metrics:

- kube_pod_container_resource_requests_cpu_cores
- kube_pod_container_resource_limits_cpu_cores
- kube_pod_container_resource_requests_memory_bytes
- kube_pod_container_resource_limits_memory_bytes
- kube_pod_container_resource_requests_nvidia_gpu_devices
- kube_pod_container_resource_limits_nvidia_gpu_devices



### Configure Alertmanager

You can configure the Alertmanager using the [alertmanager](../alertmanager/README.md) chart.
