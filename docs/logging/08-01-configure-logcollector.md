---
title: Configure the log collector
type: Tutorials
---

By default, Loki comes with the [promtail](https://github.com/grafana/loki) log collector configuration. Additionally, Kyma supports other log collectors, such as [Fluent Bit](https://fluentbit.io/) you can easily configure.

Follow these steps to adjust the Loki configuration to use Fluent Bit and Elasticsearch:

1. Override the [`values.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/logging/values.yaml) file includes the defined **logcollector** parameter. See the example: 
```yaml
logcollector:
  name: fleunt-bit
```
For details on configurable parameters and overrides, see [this](/components/logging/#configuration-configuration) document.
2. Add the following configuration to the `fluent-bit-configmap.yaml` file for Fluent Bit to forward logs to Elasticsearch. 

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

>**NOTE:** Configure **{FLUENT_ELASTICSEARCH_HOST}** and  **{FLUENT_ELASTICSEARCH_PORT}**  environment variables accordingly for your Elasticsearch deployment.
