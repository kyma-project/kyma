---
title: Send notifications to Slack
type: Tutorials
---
This tutorial shows you how to configure Alertmanager to send notifications. Alertmanager supports several [notification receivers](https://prometheus.io/docs/alerting/configuration/), however, this tutorial addresses Slack.

## Prerequisites

1. This  tutorial is a follow-up of the [observe application metrics](/components/monitoring/#tutorials-observe-application-metrics) and the  [define alerting rules](https://kyma-project.io/docs/master/components/monitoring/#tutorials-define-alerting-rules) tutorials that use the `monitoring-custom-metrics` example. Follow this tutorial to deploy the `sample-metrics-8081` service which exposes the `cpu_temperature_celsius` metric, and create an alert based on it. That configuration is required to complete this tutorial.
2. Install the Incoming Webhooks Slack app for Slack to receive notifications coming from third party services. Read [this](https://api.slack.com/incoming-webhooks#create_a_webhook) document to find out more about the app.

## Steps

1. The configuration for notification receivers is located in [this](https://github.com/kyma-project/kyma/blob/master/resources/monitoring/charts/alertmanager/templates/alertmanager.config.yaml) template, which uses the values provided in the in the [values.yaml](./values.yaml) file. 

**TIP**: To avoid confusion, use the configuration parameters for `route` and then group the receivers under the label `routes`.

```yaml
{{ define "alertmanager.yaml.tpl" }}
global:
  resolve_timeout: 5m
route:
  receiver: 'null'
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 5m
  group_by: ['cluster','pod','job','alertname']
  # All alerts that do not match the following child routes
  # remain at the root Node and are dispatched to the `default-receiver`.
  routes:
  - receiver: 'null'
    match:
      alertname: CPUTempHigh
  - receiver: "victorOps"
    continue: true # If set to `false`, it stops after the first matching.
    match_re:
      severity: critical
  - receiver: "slack"
    continue: true # If set to `false`, it stops after the first matching.
    match_re:
      severity: warning|critical
receivers:
- name: 'null'
- name: "victorOps"
  victorops_configs:
  - api_key: {{ .Values.global.alertTools.credentials.victorOps.apikey | quote }}
    send_resolved: true
    api_url: <victor-ops-url>
    routing_key: {{ .Values.global.alertTools.credentials.victorOps.routingkey | quote }}
    state_message: 'Alert: {{`{{ .CommonLabels.alertname }}`}}. Summary:{{`{{ .CommonAnnotations.summary }}`}}. RawData: {{`{{ .CommonLabels }}`}}'
- name: "slack"
  slack_configs:
  - channel: {{ .Values.global.alertTools.credentials.slack.channel | quote }}
    send_resolved: true
    api_url: {{ .Values.global.alertTools.credentials.slack.apiurl | quote }}
    icon_emoji: ":ghost:"
    title: '[{{`{{ .Status | toUpper }}`}}{{`{{ if eq .Status "firing" }}`}}:{{`{{ .Alerts.Firing | len }}`}}{{`{{ end }}`}}] Monitoring Event Notification'
    text: "<!channel> \nsummary: {{`{{ .CommonAnnotations.summary }}`}}\ndescription: {{`{{ .CommonAnnotations.description }}`}}"
{{ end }}
```

3. Configure the global alert notifications settings using the following parameters:

| Parameter | Description |
|-----------|-------------|
| **route** | A route block defines a node in a routing tree and its children. Its optional configuration parameters are inherited from its parent node if not set. | 
| **route.routes** | Child routes.  |
| **routes.receiver** | Receiver is a named configuration of one or more notification integrations.  |
| **receivers** | A list of configured notification receivers.This configuration enables the **receivers**, VictorOps and Slack to receive alerts fired by Prometheus rules. |

For details on Alertmanager chart configuration and parameters see [this](components/monitoring/#configuration-alertmanager-sub-chart) document.

4. Define a Secret to [override](../../../../docs/kyma/05-03-overrides.md) the default values.

```yaml
apiVersion: v1
kind: Secret
metadata:
 name: monitoring-config-overrides
 namespace: kyma-installer
 labels:
    kyma-project.io/installation: ""
    installer: overrides
    component: monitoring
type: Opaque
stringData:
    global.alertTools.credentials.slack.channel: "{CHANNEL_NAME}"
    global.alertTools.credentials.slack.apiurl: "{HOOK_ENDPOINT}"
```
Use the following parameters:

| Parameter | Description | Example|
|-----------|-------------|---------|
| **global.alertTools.credentials.slack.channel** | Refers to the Slack channel which receives notifications on new alerts. | `monitoring-alerts`|
| **global.alertTools.credentials.slack.apiurl** | Specifies the URL endpoint which sends alerts triggered by Prometheus rules. | `https://hooks.slack.com/services/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`|

4. Proceed with Kyma installation or, if your are using the runtime already, apply the override by running:
```
kubectl label installation/kyma-installation action=install
```
