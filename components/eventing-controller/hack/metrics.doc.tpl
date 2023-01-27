---
title: Kyma Eventing Metrics
---

Kyma Eventing provides several Grafana Dashboard with various [metrics](./evnt-02-eventing-metrics.md), so you can monitor statistics and other information in real time.
The metrics follow the [Prometheus naming convention](https://prometheus.io/docs/practices/naming/).

### Metrics Emitted by Eventing Publisher Proxy:

| Metric                                                          | Description                                                                                   |
|-----------------------------------------------------------------|:----------------------------------------------------------------------------------------------|
{{- range (ds "epp") }}
| **{{.Name}}** | {{.Help}} |
{{- end }}
### Metrics Emitted by Eventing Controller:

| Metric                                      | Description                                                                    |
|---------------------------------------------|:-------------------------------------------------------------------------------|
{{- range (ds "ec") }}
| **{{.Name}}** | {{.Help}} |
{{- end }}

### Metrics Emitted by NATS Exporter:

The [Prometheus NATS Exporter](https://github.com/nats-io/prometheus-nats-exporter) also emits metrics that you can monitor. Learn more about [NATS Monitoring](https://docs.nats.io/running-a-nats-service/configuration/monitoring#jetstream-information).  
 
