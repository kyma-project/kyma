### Alertmanager

In Kyma part of the configuration related to the Alertmanager is in this chart.

#### Secret configuration

Alertmanager instances require a secret resource to be named with the following format alertmanager-{ALERTMANAGER_NAME}.

In Kyma, the name of the Alertmanager is defined by ```name: {{ .Release.Name }}```. The secret is ```name: alertmanager-{{ .Release.Name }}```. The name of the config file is alertmanager.yaml.

```yaml
apiVersion: v1
kind: Secret
metadata:
  labels:
    alertmanager: {{ .Release.Name }}
    app: {{ template "alertmanager.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: alertmanager-{{ .Release.Name }}
data:
  alertmanager.yaml: |-
    {{ include "alertmanager.yaml.tpl" . | b64enc }}
{{- range $key, $val := .Values.templateFiles }}
  {{ $key }}: {{ $val | b64enc | quote }}
{{- end }}
```

Secret `data:` is an encoded ```alertmanager.yaml``` file which contains all the configuration for alerting notifications.

The next section explains how to configure Alertmanager for enabling alerting notifications.


#### Alertmanager configuration - alertmanager.yaml

The template
[kyma/resources/core/charts/monitoring/charts/alertmanager/templates/alertmanager.config.yaml](templates/secret.yaml) pre-configure two simple receiver to handle alert in **VictorOps and Slack**.

This yaml file pre-configure two simple receivers to handle alert in **VictorOps and Slack**.

To avoid confusion, use optional configuration parameters for ```route:``` and then group the receivers under the label ```routes:```

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
  # will remain at the root node and be dispatched to 'default-receiver'
  routes:
  - receiver: 'null'
    match:
      alertname: DeadMansSwitch
  - receiver: "victorOps"
    continue: true # If continue: is set to false it will stop after the first matching.
    match_re:
      severity: critical
  - receiver: "slack"
    continue: true # If continue: is set to false it will stop after the first matching.
    match_re:
      severity: warning|critical
receivers:
- name: 'null'
- name: "victorOps"
  victorops_configs:
  - api_key: {{ .Values.global.alertTools.credentials.victorOps.apikey | quote }}
    send_resolved: true
    api_url: https://alert.victorops.com/integrations/generic/20131114/alert/
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
**A Quick explanation**
* ```route:``` A route block defines a node in a routing tree and its children. Its optional configuration parameters are inherited from its parent node if not set.
* ```routes:``` Child routes.
* ```receiver:``` Receiver is a named configuration of one or more notification integrations.
* ```receiver:``` A list of configured notification receivers.

What this configuration provides is to enable the ```receivers:```, "victorOps" and "slack" for receiving alerts fired by Prometheus rules.

In order to enable alert notifications for the receivers above, four parameters need to be configured.

* ```api_key:``` defines the team Api key in VictorOps
* ```routing_key:``` defines the team routing key in VictorOps
* ```channel:```  the Slack channel to receive the alerts notifications
* ```api_url:``` The url endpoint containing to send the alerts.

As it was mentioned at first only part of the configuration is located in this chart, that is why all of four parameters values are got it from the template, ```{{ .Values.global.alertTools.credentials... }}```, and these values are configured in
[kyma/resources/core/values.yaml](../../../../values.yaml). There we have:


```yaml
global:
  #... some other configuration here
  #Alerting tools credentials
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
```

The main reason for keeping this configuration as a ```global:``` is because these parameters might be replaced with values configured during a cluster build and taken from the final enviromente variables during the Kyma installation.

**References**
- [VictorOps-Prometheus Integration Guide](https://help.victorops.com/knowledge-base/victorops-prometheus-integration/)
- [Prometheus Alerting configuration](https://prometheus.io/docs/alerting/configuration/)
- [Prometheus Alerting notification examples](https://prometheus.io/docs/alerting/notification_examples/)
- [Slack Incoming WebHooks](https://slack.com/apps/A0F7XDUAZ-incoming-webhooks)
- [Slack-API Legacy custom integrations](https://api.slack.com/custom-integrations)


### Create Alert Rules

In Kyma all the configuration related to alert rules is in the chart [alert-rules](../alert-rules/README.md)
