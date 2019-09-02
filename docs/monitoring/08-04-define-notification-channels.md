---
title: Define notification channels
type: Tutorials
---
This tutorial shows you how to set up notification channels for your alerts.

## Alertmanager configuration - alertmanager.yaml

This section explains how to configure Alertmanager to enable alerting notifications. [This](templates/alertmanager.config.yaml) template pre-configures two simple receivers to handle alerts in VictorOps and Slack.

This yaml file pre-configures two simple receivers to handle alerts in VictorOps and Slack.

To avoid confusion, use optional configuration parameters for `route:` and then group the receivers under the label ```routes:```

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

Configure your alert rule using the following parameters:

| Parameter | Description |
|-----------|-------------|
| **route** | A route block defines a node in a routing tree and its children. Its optional configuration parameters are inherited from its parent node if not set. | 
| **route.routes** | Child routes.  |
| **routes.receiver** | Receiver is a named configuration of one or more notification integrations.  |
| **receivers** | A list of configured notification receivers.|


This configuration enables the **receivers**, VictorOps, and Slack to receive alerts fired by Prometheus rules.

In order to enable alert notifications for the receivers, configure these parameters:

VictorOps:

**api_key** defines the team Api key in VictorOps.
**routing_key** defines the team routing key in VictorOps.
**api_url** is the URL endpoint which sends the alerts.

Slack: 
**channel** refers to the Slack channel which receives the alerts notifications.
**api_url** is the URL endpoint which sends the alerts.

Only a part of the configuration is located in this chart. All of the four parameters' values are passed as [overrides](../../../../docs/kyma/05-03-overrides.md) during the installation.
The overrides must have `global.alertTools.credentials` prefix and are used to configure the following fragment of [values.yaml](./values.yaml) file:

```yaml
global:
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
```

The main reason to keep this configuration as **global** is that these parameters might be replaced with values configured during the cluster build and taken from the final environment variables during the Kyma installation.
