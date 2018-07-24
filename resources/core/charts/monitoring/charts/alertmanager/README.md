### Alertmanager

In Kyma all the configuration related to the Alertmanager is in this chart.

#### Secret configuration

Alertmanager instances require the secret resource naming to follow the format alertmanager-{ALERTMANAGER_NAME}.

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
  alertmanager.yaml: {{ toYaml .Values.config | b64enc | quote }}
{{- range $key, $val := .Values.templateFiles }}
  {{ $key }}: {{ $val | b64enc | quote }}
{{- end }}
```

The next section explains the Alertmanager configuration.

#### Alertmanager configuration - alertmanager.yaml

[kyma/resources/core/charts/monitoring/charts/alertmanager/values.yaml](values.yaml) pre-configure two simple receiver to handle alert in **VictorOps and Slack**.

To avoid confusion, use optional configuration parameters for ```route:``` and then group the receivers under the label ```routes:```

```yaml
config:
  global:
    resolve_timeout: 5m
  route:
    receiver: 'null'
    group_wait: 30s
    group_interval: 5m
    repeat_interval: 1h # change to 10m to test
    group_by: ['cluster','pod','job','alertname']
    # All alerts that do not match the following child routes
    # will remain at the root node and be dispatched to 'default-receiver'
    routes:
    - receiver: 'null'
      match:
        alertname: DeadMansSwitch
    # - receiver: 'team-YOUR-TEAM-victorOps'
    #   continue: true # If continue: is set to false it will stop after the first matching.
    #   match:
    #     alertname: PodNotRunning
    # - receiver: 'team-YOUR-TEAM-slack'
    #   continue: true # If continue: is set to false it will stop after the first matching.
    #   match:
    #     alertname: PodNotRunning
  receivers:
  - name: 'null'
  # - name: 'team-YOUR-TEAM-victorOps'
  #   victorops_configs:
  #     - api_key: API_VICTOROPS_TOKEN
  #       send_resolved: true
  #       api_url: https://alert.victorops.com/integrations/generic/20131114/alert/
  #       routing_key: YOUR-TEAM-ROUTING-KEY
  #       state_message: 'Alert: {{ .CommonLabels.alertname }}. Summary:{{ .CommonAnnotations.summary }}. RawData: {{ .CommonLabels }}'
  # - name: 'team-YOUR-TEAM-slack'
  #   slack_configs:
  #   - channel: '#YOU_CHANNEL'
  #     send_resolved: true
  #     api_url: https://hooks.slack.com/services/XXXXXXXXX/XXXXXXXXX/XXXXXXXXXXXXXXXXXX #Your slack Webhook URL
  #     icon_emoji: ":ghost:"
  #     title: '[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] Monitoring Event Notification'
  #     text: "<!channel> \nsummary: {{ .CommonAnnotations.summary }}\ndescription: {{ .CommonAnnotations.description }}"
```
**A Quick explanation**
* ```route:``` A route block defines a node in a routing tree and its children. Its optional configuration parameters are inherited from its parent node if not set.
* ```routes:``` Child routes.
* ```receiver:``` Receiver is a named configuration of one or more notification integrations.
* ```receiver:``` A list of configured notification receivers.

**References**
- [VictorOps-Prometheus Integration Guide](https://help.victorops.com/knowledge-base/victorops-prometheus-integration/)
- [Prometheus Alerting configuration](https://prometheus.io/docs/alerting/configuration/)
- [Prometheus Alerting notification examples](https://prometheus.io/docs/alerting/notification_examples/)
- [Slack Incoming WebHooks](https://slack.com/apps/A0F7XDUAZ-incoming-webhooks)
- [Slack-API Legacy custom integrations](https://api.slack.com/custom-integrations)


### Create Alert Rules

In Kyma all the configuration related to alert rules is in the chart [alert-rules](../alert-rules/README.md)
