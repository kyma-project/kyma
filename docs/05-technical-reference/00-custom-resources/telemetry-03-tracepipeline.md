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
        value: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:4317
status:
  conditions:
  - lastTransitionTime: "2022-12-13T14:33:27Z"
    reason: OpenTelemetryDeploymentNotReady
    type: Pending
  - lastTransitionTime: "2022-12-13T14:33:28Z"
    reason: OpenTelemetryDeploymentReady
    type: Running
```

For further TracePipeline examples, see the [samples](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/config/samples) directory.

## Custom resource parameters

### spec attribute

For details, see the [TracePipeline specification file](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/apis/telemetry/v1alpha1/tracepipeline_types.go).

| Parameter | Type | Description |
|---|---|---|
| output | object | Defines a destination for shipping trace data. Only one can be defined per pipeline.
| output.otlp | object | Configures the underlying Otel Collector with an [OTLP exporter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/otlpexporter/README.md). If you switch `protocol`to `http`, an [OTLP HTTP exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter) is used. |
| output.otlp.protocol | string | Use either GRPC or HTTP protocol. Default is GRPC. |
| output.otlp.endpoint | object | Configures the endpoint of the destination backend in format `<scheme>://<host>:<port>` where host and port are mandatory. |
| output.otlp.endpoint.value | string | Endpoint taken from a static value |
| output.otlp.endpoint.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| output.otlp.authentication | object | Configures the authentication mechanism for the destination. |
| output.otlp.authentication.basic | object | Activates `Basic` authentication for the destination providing relevant Secrets. |
| output.otlp.authentication.basic.password | object | Configures the password to be used for `Basic` authentication. |
| output.otlp.authentication.basic.password.value | string | Password as plain text provided as static value. Do not use in production, as it does not satisfy standards for secret handling. Use the `valueFrom.secretKeyRef` instead. |
| output.otlp.authentication.basic.password.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`.|
| output.otlp.authentication.basic.user | object | Configures the username to be used for `Basic` authentication. |
| output.otlp.authentication.basic.user.value | string | Username as plain text provided as static value. |
| output.otlp.authentication.basic.user.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |

### status attribute

For details, see the [TracePipeline specification file](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/apis/telemetry/v1alpha1/tracepipeline_types.go).

| Parameter | Type | Description |
|---|---|---|
| conditions | []object | An array of conditions describing the status of the pipeline.
| conditions[].lastTransitionTime | []object | An array of conditions describing the status of the pipeline.
| conditions[].reason | []object | An array of conditions describing the status of the pipeline.
| conditions[].type | enum | The possible transition types are:<br>- `Running`: The instance is ready and usable.<br>- `Pending`: The pipeline is being activated. |
