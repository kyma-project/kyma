# Alert Rules

## Overview

Kyma uses Prometheus alert rules to monitor the health of resources. Use this chart to configure alert rules.

## Details

### Alert rules

You can define the following alert rules:

- Alert when a Pod is not running

    The Alertmanager sends out alerts when one of the Pods is not running in the `kyma-system`, `kyma-integration`, `istio-system`, `kube-public`, or `kube-system` Namespace.

- Monitor Persistent Volume Claims (PVCs)

    The Alertmanager sends out alerts when the PVC exceeds 90% for the following system Namespaces: `kyma-system`, `kyma-integration`, `heptio-ark`, `istio-system`, `kube-public`, or `kube-system`. To avoid this, increase the capacity of the PVC.

-  Monitor CPU Usage

    The Alertmanager sends out alerts when the CPU usage exceeds 90% for Pods in the `kyma-system` Namespace. Add the `alertcpu: "yes"` label to Pods to make sure the rule activates.

- Monitor memory usage

    The Alertmanager triggers the rule when memory usage exceeds 90% for Pods in the `kyma-system` Namespace. Add the `alertmem: "yes"` label to Pods to make sure the rule activates.

### Create alert rules

Prometheus uses the  **spec.ruleSelector** label selector to identify ConfigMaps which include Prometheus rule definitions. 

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

1. Use the [prometheusRule template](./templates/kyma-rules.yaml) which contains the sample configuration for an alert rule.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kyma.rules
  labels:
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    prometheus: {{ .Release.Name }}
    release: {{ .Release.Name }}
    role: alert-rules
    app: kyma.rules
spec:
  groups:
  - name: pod-not-running-rule
    rules:
    - alert: SystemPodNotRunning
      expr: sum(kube_pod_container_status_running { namespace=~"kyma-.*|kube-.*|istio-.*|natss", pod!~"(test.*)|((dummy|sample)-.*)|(.*(docs|backup|test)-.*)|(.*-(tests|dummy))" } == 0 )by (pod,namespace) * on(pod, namespace) (kube_pod_status_phase{phase="Succeeded"} != 1)
      for: 60s
      labels:
        severity: critical
      annotations:
        description: "{{`{{ $labels.namespace }}`}}/{{`{{ $labels.pod }}`}} is not running"
        summary: "{{`{{ $labels.pod }}`}} is not running"
```

2. Label the PrometheusRule with `role: alert-rules`.
3. Add the name of a Prometheus object in `prometheus: {{ .Release.Name }}`.

The rule definition includes the following parameters:

- **alert:** is the valid metric name of the alert.
- **expr:** defines the PromQL expression to evaluate, using Kubernetes [functions](https://prometheus.io/docs/prometheus/latest/querying/functions/) and [metrics](https://github.com/kubernetes/kube-state-metrics/blob/master/Documentation/pod-metrics.md). In the example, the `kube_pod_container_status_running` Pod metric is used to check if the `sample-metrics` Pod is running in the `default` Namespace.
* **for:**  is a time period during which alerts are returned.
* **description:** is an annotation used to enrich alert details.
* **summary:** is an annotation used to enrich alert details.


### Configure Alertmanager

You can configure the Alertmanager using the [alertmanager](../alertmanager/README.md) chart.
