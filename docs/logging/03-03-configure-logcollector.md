---
title: Log collector configuration examples
type: Details
---

## Storage
By default, Loki comes with the [promtail](https://github.com/grafana/loki) log collector configuration. Additionally, kyma supports other log collector, such as [fluent-bit](https://fluentbit.io/).

This is an example of Loki configuration using fluent-bit and elesticsearch:

File `values.yaml` from kyma logging helm chart have to be configured like following example:
```yaml
logcollector:
  name: promtailfleunt-bit
```

Add following configuration to the `fluent-bit-configmap.yaml` to configure fluent-bit to forward logs to the elastic search. Environment variables `${FLUENT_ELASTICSEARCH_HOST}` and `${FLUENT_ELASTICSEARCH_PORT}` should be configured accordingly to your elastic search deployment.
```yaml
    output-elasticsearch.conf: |
        [OUTPUT]
            Name            es
            Match           *
            Host            ${FLUENT_ELASTICSEARCH_HOST}
            Port            ${FLUENT_ELASTICSEARCH_PORT}
            Logstash_Format On
            Replace_Dots    On
            Retry_Limit     False
```