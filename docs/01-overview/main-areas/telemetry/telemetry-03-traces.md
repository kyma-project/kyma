---
title: Telemetry - Traces
---

Observability tools should clearly show the big picture, no matter if you're monitoring just a few or many components. In a cloud-native microservice architecture, a user request often flows through dozens of different microservices. Logging and monitoring tools help to track the request's path. However, they treat each component or microservice in isolation. This individual treatment results in operational issues.

[Distributed tracing](https://opentelemetry.io/docs/concepts/observability-primer/#understanding-distributed-tracing) charts out the transactions in cloud-native systems, helping you to understand the application behavior and relations between the frontend actions and backend implementation.

The diagram shows how distributed tracing helps to track the request path:

![Distributed tracing](./assets/tracing-intro.drawio.svg)

## Prerequisites

For recording a distributed trace complete, it is [essential](https://www.w3.org/TR/trace-context/#problem-statement) that every involved component is at least propagating the trace context. In Kyma, all components involved in users requests are supporting the [W3C Trace Context protocol](https://www.w3.org/TR/trace-context), which is a vendor-neutral protocol getting more and more supported by all kind of vendors and tools. The involved Kyma components are mainly Istio, Serverless and Eventing.

With that, your application must propagate as well the W3C Trace Context for any user related user activity. This can be achieved easily by using the [Open Telemetry SDKs](https://opentelemetry.io/docs/instrumentation/) available for all common programming languages. If an application is following that guidance and is part of the Istio Service Mesh, it already will be outlined with dedicated span data in the trace data collected by the Kyma telemetry setup.

Furthermore, an application should enrich a trace with additional span data and send these data to the cluster-central telemetry services. That as well, can be achieved with the help of the mentioned [Open Telemetry SDKs](https://opentelemetry.io/docs/instrumentation/).

## Architecture

The telemetry module provides an in-cluster central deployment of an [Otel Collector](https://opentelemetry.io/docs/collector/). The collector is exposing endpoints for the OTLP protocol for GRPC and HTTP based communication via a dedicates Service `telemetry-otlp-traces`. Here, all Kyma components and users applications should send trace data to.

![Architecture](./assets/tracing-arch.drawio.svg)

### Otel Collector

### Telemetry Operator

### Pipelines

## Setting up a TracePipeline

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: jaeger
spec:
  output:
    otlp:
      endpoint:
        value: tracing-jaeger-collector.kyma-system.svc.cluster.local:4317
```

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: jaeger
spec:
  output:
    otlp:
      protocol: http
      endpoint:
        value: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:4318
```

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
      authentication:
        basic:
          user:
            value: myUser
          password:
            value: myPwd
```

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: jaeger
spec:
  output:
    otlp:
      endpoint:
        valueFrom:
            secretKeyRef:
                name: backend
                namespace: default
                key: endpoint
      authentication:
        basic:
          user:
            valueFrom:
              secretKeyRef:
                 name: backend
                 namespace: default
                 key: user
          password:
            valueFrom:
              secretKeyRef:
                 name: backend
                 namespace: default
                 key: password
```

## Parameters


### TracePipeline.spec attribute

For details, see the [TracePipeline specification file](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/apis/telemetry/v1alpha1/tracepipeline_types.go).

| Parameter | Type | Description |
|---|---|---|
| output | object | Defines a destination for shipping trace data. Only one can be defined per pipeline.
| output.otlp | object | Configures the underlying Otel Collector with an [OTLP exporter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/otlpexporter/README.md). By switching the `protocol`to `http` a [OTLP HTTP exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter) is used. |
| output.otlp.protocol | string | Use either GRPC or HTTP protocol. Default is GRPC. |
| output.otlp.endpoint | object | Configures the endpoint of the destination backend in format `<scheme>://<host>:<port>` where host and port are mandatory. |
| output.otlp.endpoint.value | string | Endpoint taken from a static value |
| output.otlp.endpoint.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| output.otlp.authentication | object | Configures the authentication mechanism for the destination. |
| output.otlp.authentication.basic | object | Activates `Basic` authentication for the destination providing relevant secrets. |
| output.otlp.authentication.basic.password | object | Configures the password to be used for `Basic` authentication. |
| output.otlp.authentication.basic.password.value | string | Password as plain text provided as static value. Do not use in production, as it does not satisfy standards for secret handling. Use the `valueFrom.secretKeyRef` instead. |
| output.otlp.authentication.basic.password.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`.|
| output.otlp.authentication.basic.user | object | Configures the username to be used for `Basic` authentication. |
| output.otlp.authentication.basic.user.value | string | Username as plain text provided as static value. |
| output.otlp.authentication.basic.user.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |

### TracePipeline.status attribute

For details, see the [TracePipeline specification file](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/apis/telemetry/v1alpha1/tracepipeline_types.go).

| Parameter | Type | Description |
|---|---|---|
| conditions | []object | An array of conditions describing the status of the pipeline.
| conditions[].lastTransitionTime | []object | An array of conditions describing the status of the pipeline.
| conditions[].reason | []object | An array of conditions describing the status of the pipeline.
| conditions[].type | enum | The possible transition types are:<br>- Running: The instance is ready and usable.<br>- Pending: The pipeline is being activated. |

## Trace processing

## Limitations

The collector setup is designed using the following assumptions:
- 

Out of that the following limitations are resulting:

### Throughput

### Unavailability of output

