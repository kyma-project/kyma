---
title: Alertmanager
type: Details
---

Alertmanager receives and manages alerts coming from Prometheus. It can then forward the notifications about fired alerts to specific channels, such as Slack or VictorOps. 

## Alertmanager configuration

Use the following files to configure and manage Alertmanager:

* `alertmanager.yaml` which deploys the Alertmananger Pod. 
* `alertmanager.config.yaml` which you can use to define core Alertmanager configuration and alerting channels. For details on configuration elements, see [this](https://prometheus.io/docs/alerting/configuration/) document.
* `alertmanager.rules` which lists alerting rules used to monitor Alertmanager's health.

Additionally, Alertmanager instances require a [Secret](../../resources/monitoring/charts/alertmanager/templates/secret.yaml) resource which contains the encoded `alertmanager.yaml.tpl` file. This Secret is picked up during Pod deployment and mounted as `alertmanager.config.yaml`, which allows you to configure alert settings and notifications.

The Secret resource looks as follows:

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
To configure the alerts and be able to forward them to different channels, define the parameters: 

| Parameter | Description |
|-----------|-------------|
| **name** | Specifies the name of the Secret. The name must follow the `alertmanager-{ALERTMANAGER_NAME}` format. |
| **data** | Contains the encoded `alertmanager.yaml.tpl` file which contains all the configuration for alerting notifications provided in the [`alertmanager.config.yaml`](../../resources/monitoring/charts/alertmanager/templates/alertmanager.config.yaml) file.|


## Alerting rules

Kyma comes with a set of alerting rules provided out of the box. You can find them [here](../../resources/monitoring/charts/alert-rules/templates).
These rules provide alerting configuration for logging, webapps, rest services, and custom Kyma rules. 

You can also define your own alerting rule. To learn how, see [this](components/monitoring/#tutorials-define-alerting-rules) tutorial.
