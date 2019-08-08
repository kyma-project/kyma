---
title: Alertmanager
type: Details
---

Alertmanager receives and manages alerts coming from Prometheus. It can then forward the fired alerts to spefied channels, such as Slack. 

## Alertmanager configuration

The Alertmanager configuration is located in the `alertmanager.config.yaml`. For details on configuration elements, see [this](https://prometheus.io/docs/alerting/configuration/) document.

Alertmanager instances require a [Secret](../../resources/monitoring/charts/alertmanager/templates/secret.yaml) resource, whose name follows the `alertmanager-{ALERTMANAGER_NAME}` format.

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

>**NOTE**: The **data** property contains an encoded `alertmanager.yaml` file which contains all the configuration for alerting notifications.


## Alerting rules

Kyma comes with a set of alerting)rules provided out of the box. You can find them [here](../../resources/monitoring/charts/alert-rules/templates)

The rules include:
* logging rules
* rest services rules
* webapps rules
* custom Kyma alert rules
