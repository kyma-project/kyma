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

1. An end-to-end request is triggered and populates across the distributed application. Every involved component is propagating the trace context using the [w3c-tracecontext](https://www.w3.org/TR/trace-context/) protocol.
1. The involved components which have contributed a new span to the trace send the related span data to the trace collector using the `telemetry-otlp-traces` service. The communication happens on base of the [OTLP](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md) protocol either using GRPC ot HTTP.
1. the trace collector enriches the span data with relevant metadata typical for sources running on kubernetes, like pod identifiers.
1. Via a `LogPipeline` resource, the trace collector got configured with a target backend
1. The backend can run either in-cluster
1. or outer-cluster having a proper way of authentication in place.
1. The trace data can be consumed via the backend system

### Otel Collector
The Otel Collector comes with a [concept](https://opentelemetry.io/docs/collector/configuration/) of pipelines consisting of receivers, processors and exporters with which you can flexible plug pipelines together. Kymas TracePipeline is providing you with a hardened setup of an Otel Collector and is also abstracting the underlying pipeline concept. The benefits of having that abstraction in place are:
- Supportability - all features are tested and supported
- Migratability - smooth migration experiences when switching underlying technologies or architectures
- Native K8S support - API provided by Kyma will allow easy integration with Secrets for example served by the SAP BTP Operator. The Telemetry Operator will take care of the full lifecycle.
- Focus - the user don't need to understand underlying concepts

The downside is that only a limited set of features is available. Here, you can opt-out at any time by bringing your own collector setup. The current feature set focusses on providing the full configurability of backends integrated by OTLP. As a next step, meaningful filter options will be provided. Especially head and tail based sampling configurations.

### Telemetry Operator
The TracePipeline resource is managed by the Telemetry Operator, a typical Kubernetes operator responsible for managing the custom parts of the Otel Collector configuration.

![Operator resources](./assets/tracing-resources.drawio.svg)

The Telemetry Operator watches all TracePipeline resources and related Secrets. Whenever the configuration changes, it validates the configuration and generates a new configuration for the Otel Collector, where a ConfigMaps for the configuration is generated. Furthermore, referenced Secrets are copied into one Secret that is mounted to the Collector as well.
Furthermore, the operator manages the full lifecycle of the Otel Collector Deployment itself. With that, it only gets deployed when there is an actual TracePipeline defined. At anytime you can opt-out of using the tracing feature by not specifying a TracePipeline.

## Setting up a TracePipeline

In the following a typical setup of a TracePipeline gets discussed. For a overview of all available attributes please have a look at the [reference document](./../../../05-technical-reference/00-custom-resources/telemetry-03-tracepipeline.md)

### 1. Create a TracePipeline with an output
1. To ship traces to a new OTLP output, create a resource file of kind TracePipeline:

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
```

That will configure the underlying Otel Collector with a pipeline for traces. The receiver of the pipeline will be of type OTLP and be accessible via the `telemetry-otlp-traces` service. As exporter an `otlp` or an `otlphttp` exporter gets used, dependent on the configured protocol.

2. To create the instance, apply the resource file in your cluster.
    ```bash
    kubectl apply -f path/to/my-trace-pipeline.yaml
    ```

3. Check that the status of the TracePipeline in your cluster is `Ready`:
    ```bash
    kubectl get logpipeline
    NAME              STATUS    AGE
    http-backend      Ready     44s
    ```

### 2. Switch the protocol to HTTP

To use the HTTP protocol instead of the default GRPC, use the `protocol` attribute and assure that the proper port gets configured as part of the endpoint. Typically port 4317 is used for GRPC and port 4318 for HTTP.
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

### Step 3: Add authentication details

To integrate with external systems you need to configure authentication details. At the moment only Basic Authentication is supported. A more general ton based authentication will be supported [soon](https://github.com/kyma-project/kyma/issues/16258).

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

### Step 4: Add authentication details from Secrets

Integrations into external systems usually need authentication details dealing with sensitive data. To handle that data properly in Secrets, the TracePipeline supports the reference of Secrets.

Using the **valueFrom** attribute, you can map Secret keys as in the following example:

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

The related Secret must fulfill the referenced name and Namespace, and contain the mapped key as in the following example:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: backend
  namespace: default
stringData:
  endpoint: https://myhost:4317
  user: myUser
  password: XXX
```

## Limitations

The trace collector setup is designed using the following assumptions:
- The collector has no autoscaling options yet and has a limited resource setup of 1 CPU and 1 GiB Memory
- Batching is enabled and a batch will contain up to 512 Spans/batch
- An unavailability of a destination must be survived for 5 minutes without direct loss of trace data
- An average span consists of 40 attributes mit 64 character length

Out of that the following limitations are resulting:
### Throughput
The maximum throughput is 4200 span/sec ~= 15.000.000 spans/hour. If more data needs to be ingested it might result in a refusal of more data.

### Unavailability of output
The destination can be unavailable up to 5 minutes, so a retry for data will be tried up to 5min and then data gets dropped.

### No guaranteed delivery
The used buffers are volatile and will loose the data on a crash of the ote-collector instance.

### Single TracePipeline support

Only one TracePipeline resource at a time is supported at the moment.
