---
title: Log collector configuration examples
type: Details
---

## Storage
By default, Loki comes with the [promtail](https://github.com/grafana/loki) log collector configuration. Additionally, Kyma supports other log collectors, such as [Fluent Bit](https://fluentbit.io/) you can easily configure.

Follow these steps to adjust the Loki configuration to use Fluent Bit and Elasticsearch:

File `values.yaml` from kyma logging helm chart have to be configured like following example:
```yaml
logcollector:
  name: promtailfleunt-bit
```

2. Add the following configuration to the `fluent-bit-configmap.yaml` file for Fluent Bit to forward logs to Elasticsearch. 
>**NOTE:** Configure **{FLUENT_ELASTICSEARCH_HOST}** and  **{FLUENT_ELASTICSEARCH_PORT}**  environment variables accordingly for your Elasticsearch deployment.
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
