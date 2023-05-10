---
title: Telemetry - TracePipeline
---

The `tracepipeline.telemetry.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to filter and ship trace data in Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd tracepipeline.telemetry.kyma-project.io -o yaml
```

## Sample custom resource

The following TracePipeline object defines a pipeline integrating into the local Jaeger instance.

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: jaeger
spec:
  output:
    otlp:
      endpoint:
        value: http://jaeger-collector.jaeger.svc.cluster.local:4317
status:
  conditions:
  - lastTransitionTime: "2022-12-13T14:33:27Z"
    reason: OpenTelemetryDeploymentNotReady
    type: Pending
  - lastTransitionTime: "2022-12-13T14:33:28Z"
    reason: OpenTelemetryDeploymentReady
    type: Running
```

For further TracePipeline examples, see the [samples](https://github.com/kyma-project/telemetry-manager/tree/main/config/samples) directory.

## Custom resource parameters

For details, see the [TracePipeline specification file](https://github.com/kyma-project/telemetry-manager/blob/main/apis/telemetry/v1alpha1/tracepipeline_types.go).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
### TracePipeline.telemetry.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **output** (required) | object | Defines a destination for shipping trace data. Only one can be defined per pipeline. |
| **output.<wbr>otlp** (required) | object | Configures the underlying Otel Collector with an [OTLP exporter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/otlpexporter/README.md). If you switch `protocol`to `http`, an [OTLP HTTP exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter) is used. |
| **output.<wbr>otlp.<wbr>authentication**  | object | Defines authentication options for the OTLP output |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic**  | object | Activates `Basic` authentication for the destination providing relevant Secrets. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password** (required) | object | Contains the basic auth password or a Secret reference. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password.<wbr>value**  | string | Value that can contain references to Secret values. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password.<wbr>valueFrom**  | object |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password.<wbr>valueFrom.<wbr>secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>key**  | string |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>name**  | string |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>password.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>namespace**  | string |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user** (required) | object | Contains the basic auth username or a Secret reference. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user.<wbr>value**  | string | Value that can contain references to Secret values. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user.<wbr>valueFrom**  | object |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user.<wbr>valueFrom.<wbr>secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>key**  | string |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>name**  | string |  |
| **output.<wbr>otlp.<wbr>authentication.<wbr>basic.<wbr>user.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>namespace**  | string |  |
| **output.<wbr>otlp.<wbr>endpoint** (required) | object | Defines the host and port (<host>:<port>) of an OTLP endpoint. |
| **output.<wbr>otlp.<wbr>endpoint.<wbr>value**  | string | Value that can contain references to Secret values. |
| **output.<wbr>otlp.<wbr>endpoint.<wbr>valueFrom**  | object |  |
| **output.<wbr>otlp.<wbr>endpoint.<wbr>valueFrom.<wbr>secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.<wbr>otlp.<wbr>endpoint.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>key**  | string |  |
| **output.<wbr>otlp.<wbr>endpoint.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>name**  | string |  |
| **output.<wbr>otlp.<wbr>endpoint.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>namespace**  | string |  |
| **output.<wbr>otlp.<wbr>headers**  | \[\]object | Defines custom headers to be added to outgoing HTTP or GRPC requests. |
| **output.<wbr>otlp.<wbr>headers.<wbr>name** (required) | string | Defines the header name. |
| **output.<wbr>otlp.<wbr>headers.<wbr>value**  | string | Value that can contain references to Secret values. |
| **output.<wbr>otlp.<wbr>headers.<wbr>valueFrom**  | object |  |
| **output.<wbr>otlp.<wbr>headers.<wbr>valueFrom.<wbr>secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.<wbr>otlp.<wbr>headers.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>key**  | string |  |
| **output.<wbr>otlp.<wbr>headers.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>name**  | string |  |
| **output.<wbr>otlp.<wbr>headers.<wbr>valueFrom.<wbr>secretKeyRef.<wbr>namespace**  | string |  |
| **output.<wbr>otlp.<wbr>protocol**  | string | Defines the OTLP protocol (http or grpc). Default is GRPC. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | An array of conditions describing the status of the pipeline. |
| **conditions.<wbr>lastTransitionTime**  | string | An array of conditions describing the status of the pipeline. |
| **conditions.<wbr>reason**  | string | An array of conditions describing the status of the pipeline. |
| **conditions.<wbr>type**  | string | The possible transition types are:<br>- `Running`: The instance is ready and usable.<br>- `Pending`: The pipeline is being activated. |

<!-- TABLE-END -->
