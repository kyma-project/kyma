---
title: Alertmanager
type: Details
---

Alertmanager receives and manages alerts coming from Prometheus. It can then forward the fired alerts to spefied channels, such as Slack. 

## Alertmanager configuration

You can configure and manage Alertmanager using the following files:

* `alertmanager.yaml` which deploys the Alertmananger Pod. 
* `alertmanager.config.yaml` which you can use to define core Alertmanager configuration and alerting channels. For detailson configuration elements, see [this](https://prometheus.io/docs/alerting/configuration/) document
* `alertmanager.rules` which lists alerting rules used to monitor Alertmanager's health

Additionally, Alertmanager instances require a [Secret](../../resources/monitoring/charts/alertmanager/templates/secret.yaml) resource. This resource provides an encoded  `alertmanager.yaml` file which you use to deploy the Pod.

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
| **name** | Specifies the name of the Secret. The name nas to follow the the `alertmanager-{ALERTMANAGER_NAME}` format. |
| **data** | Contains an encoded `alertmanager.yaml` file which contains all the configuration for alerting notifications.|


## Alerting rules

Kyma comes with a set of alerting rules provided out of the box. You can find them [here](../../resources/monitoring/charts/alert-rules/templates).
These rules provide alerting configuration for logging, webapps, rest services, and custom kyma rules. 
You can also define your own alerting rule To learn how, see [this](components/monitoring/#tutorials-define-alerting-rules) tutorial.
