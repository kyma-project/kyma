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

<!-- SKIP-WITH-ANCESTORS spec.template -->

<!-- TABLE-START -->
<!-- TracePipeline v1alpha1 -->
| Parameter         | Description                                   |
| ---------------------------------------- | ---------|
| **spec.output** | Configures the trace receiver of a TracePipeline. |
| **spec.output.otlp** | Defines an output using the OpenTelmetry protocol. |
| **spec.output.otlp.authentication** | Defines authentication options for the OTLP output |
| **spec.output.otlp.authentication.basic** | Contains credentials for HTTP basic auth |
| **spec.output.otlp.authentication.basic.password** | Contains the basic auth password or a secret reference |
| **spec.output.otlp.authentication.basic.password.value** |  |
| **spec.output.otlp.authentication.basic.password.valueFrom** |  |
| **spec.output.otlp.authentication.basic.password.valueFrom.secretKeyRef** |  |
| **spec.output.otlp.authentication.basic.password.valueFrom.secretKeyRef.key** |  |
| **spec.output.otlp.authentication.basic.password.valueFrom.secretKeyRef.name** |  |
| **spec.output.otlp.authentication.basic.password.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.otlp.authentication.basic.user** | Contains the basic auth username or a secret reference |
| **spec.output.otlp.authentication.basic.user.value** |  |
| **spec.output.otlp.authentication.basic.user.valueFrom** |  |
| **spec.output.otlp.authentication.basic.user.valueFrom.secretKeyRef** |  |
| **spec.output.otlp.authentication.basic.user.valueFrom.secretKeyRef.key** |  |
| **spec.output.otlp.authentication.basic.user.valueFrom.secretKeyRef.name** |  |
| **spec.output.otlp.authentication.basic.user.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.otlp.endpoint** | Defines the host and port (<host>:<port>) of an OTLP endpoint. |
| **spec.output.otlp.endpoint.value** |  |
| **spec.output.otlp.endpoint.valueFrom** |  |
| **spec.output.otlp.endpoint.valueFrom.secretKeyRef** |  |
| **spec.output.otlp.endpoint.valueFrom.secretKeyRef.key** |  |
| **spec.output.otlp.endpoint.valueFrom.secretKeyRef.name** |  |
| **spec.output.otlp.endpoint.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.otlp.headers** | Custom headers to be added to outgoing HTTP or GRPC requests |
| **spec.output.otlp.headers.name** | Defines the header name |
| **spec.output.otlp.headers.value** |  |
| **spec.output.otlp.headers.valueFrom** |  |
| **spec.output.otlp.headers.valueFrom.secretKeyRef** |  |
| **spec.output.otlp.headers.valueFrom.secretKeyRef.key** |  |
| **spec.output.otlp.headers.valueFrom.secretKeyRef.name** |  |
| **spec.output.otlp.headers.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.otlp.protocol** | Defines the OTLP protocol (http or grpc). |
| **status.conditions** | Contains details for the current condition of this TracePipeline |
| **status.conditions.lastTransitionTime** |  |
| **status.conditions.reason** |  |
| **status.conditions.type** |  |<!-- TABLE-END -->
