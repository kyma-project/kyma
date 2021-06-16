---
title: Tracing chart
---

To configure the Tracing chart, override the default values of its `values.yaml` file.
Learn how it works under [Configurable Parameters](./README.md).

You can set the following parameters:

## Configurable parameters

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.tracing.enabled** | Enables emitting traces for the relevant components. | true |
| **global.tracing.zipkinAddress** | Specifies the address of the Zipkin instance. | "zipkin.kyma-system:9411" |
| **jaeger.spec.resources.limits.memory** | Defines the maximum amount of memory that is available for storing traces in Jaeger. | `500Mi` |
| **jaeger.spec.strategy** | Deployment strategy to use. The possible values are either `allInOne` or `production`. | `allInOne` |
| **jaeger.spec.storage.type** | Defines storage type for span data. The possible values are `memory`, `badger`, `elasticsearch` `cassandra`. | `memory` |
| **jaeger.spec.storage.options** | Defines additional options for the storage type. | - |


