# Traces

The Telemetry module supports you in collecting all relevant trace data in a Kyma cluster, enriches them and ships them to a backend for further analysis. Kyma modules like Istio or Serverless contribute traces transparently. You can choose among multiple vendors for [OTLP-based backends](https://opentelemetry.io/ecosystem/vendors/).

## Overview

Observability tools aim to show the big picture, no matter if you're monitoring just a few or many components. In a cloud-native microservice architecture, a user request often flows through dozens of different microservices. Logging and monitoring tools help to track the request's path. However, they treat each component or microservice in isolation. This individual treatment results in operational issues.

[Distributed tracing](https://opentelemetry.io/docs/concepts/observability-primer/#understanding-distributed-tracing) charts out the transactions in cloud-native systems, helping you to understand the application behavior and relations between the frontend actions and backend implementation.

The following diagram shows how distributed tracing helps to track the request path:

![Distributed tracing](./assets/traces-intro.drawio.svg)

The Telemetry module provides a trace gateway for the shipment of traces of any container running in the Kyma runtime.

You can configure the trace gateway with external systems using runtime configuration with a dedicated Kubernetes API ([CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions)) named TracePipeline.
The Trace feature is optional. If you don't want to use it, simply don't set up a TracePipeline.

## Prerequisites

For the recording of a distributed trace, every involved component must propagate at least the trace context. For details, see [Trace Context](https://www.w3.org/TR/trace-context/#problem-statement).

- In Kyma, all modules involved in users’ requests support the [W3C Trace Context](https://www.w3.org/TR/trace-context) protocol. The involved Kyma modules are, for example, Istio, Serverless, and Eventing.
- Your application also must propagate the W3C Trace Context for any user-related activity. This can be achieved easily using the [Open Telemetry SDKs](https://opentelemetry.io/docs/instrumentation/) available for all common programming languages. If your application follows that guidance and is part of the Istio Service Mesh, it’s already outlined with dedicated span data in the trace data collected by the Kyma telemetry setup.
- Furthermore, your application must enrich a trace with additional span data and send these data to the cluster-central telemetry services. You can achieve this with [Open Telemetry SDKs](https://opentelemetry.io/docs/instrumentation/).

## Architecture

In the Kyma cluster, the Telemetry module provides a central deployment of an [OTel Collector](https://opentelemetry.io/docs/collector/) acting as a gateway. The gateway exposes endpoints to which all Kyma modules and users’ applications should send the trace data.

![Architecture](./assets/traces-arch.drawio.svg)

1. An end-to-end request is triggered and populated across the distributed application. Every involved component propagates the trace context using the [W3C Trace Context](https://www.w3.org/TR/trace-context/) protocol.
2. After contributing a new span to the trace, the involved components send the related span data to the trace gateway using the `telemetry-otlp-traces` service. The communication happens based on the [OpenTelemetry Protocol (OTLP)](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md) either using GRPC or HTTP.
3. Istio sends the related span data to the trace gateway as well.
4. The trace gateway discovers metadata that's typical for sources running on Kubernetes, like Pod identifiers, and then enriches the span data with that metadata.
5. Telemetry Manager configures the gateway according to the `TracePipeline` resource, including the target backend for the trace gateway. Also, it observes the trace flow to the backend and reports problems in the `TracePipeline` status.
6. The trace gateway sends the data to the observability system that's specified in your `TracePipeline` resource - either within the Kyma cluster, or, if authentication is set up, to an external observability backend.
7. You can analyze the trace data with your preferred backend system.

### Telemetry Manager

The TracePipeline resource is watched by Telemetry Manager, which is responsible for generating the custom parts of the OTel Collector configuration.

![Manager resources](./assets/traces-resources.drawio.svg)

1. Telemetry Manager watches all TracePipeline resources and related Secrets.
2. Furthermore, Telemetry Manager takes care of the full lifecycle of the OTel Collector Deployment itself. Only if you defined a TracePipeline, the collector is deployed.
3. Whenever the configuration changes, it validates the configuration and generates a new configuration for OTel Collector, where a ConfigMap for the configuration is generated.
4. Referenced Secrets are copied into one Secret that is mounted to the OTel Collector as well.

### Trace Gateway

In a Kyma cluster, the trace gateway is the central component to which all components can send their individual spans. The gateway collects, enriches, and dispatches the data to the configured backend. For more information, see [Telemetry Gateways](./gateways.md).

## Setting up a TracePipeline

In the following steps, you can see how to construct and deploy a typical TracePipeline. Learn more about the available [parameters and attributes](resources/04-tracepipeline.md).

### 1. Create a TracePipeline

To ship traces to a new OTLP output, create a resource of the kind `TracePipeline` and save the file (named, for example, `tracepipeline.yaml`).

This configures the underlying OTel Collector with a pipeline for traces and opens a push endpoint that is accessible with the `telemetry-otlp-traces` service. For details, see [Gateway Usage](./gateways.md#usage). The following push URLs are set up:

- GRPC: 'http://telemetry-otlp-traces.kyma-system:4317'
- HTTP: 'http://telemetry-otlp-traces.kyma-system:4318'

The default protocol for shipping the data to a backend is GRPC, but you can choose HTTP instead. Depending on the configured protocol, an `otlp` or an `otlphttp` exporter is used. Ensure that the correct port is configured as part of the endpoint.

- For GRPC, use:

  ```yaml
  apiVersion: telemetry.kyma-project.io/v1alpha1
  kind: TracePipeline
  metadata:
    name: backend
  spec:
    output:
      otlp:
        endpoint:
          value: https://backend.example.com:4317
  ```

- For HTTP, use the `protocol` attribute:

  ```yaml
  apiVersion: telemetry.kyma-project.io/v1alpha1
  kind: TracePipeline
  metadata:
    name: backend
  spec:
   output:
      otlp:
        protocol: http
        endpoint:
          value: https://backend.example.com:4318
  ```
  
### 2. Enable Istio Tracing

By default, the tracing feature of the Istio module is disabled to avoid increased network utilization if there is no TracePipeline.

To activate the Istio tracing feature with a sampling rate of 5% (for recommendations, see [Istio](#istio)), use a resource similar to the following example:

```yaml
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: tracing-default
  namespace: istio-system
spec:
  tracing:
  - providers:
    - name: "kyma-traces"
    randomSamplingPercentage: 5.00
```

### 3a. Add Authentication Details From Plain Text

To integrate with external systems, you must configure authentication  details. You can use mutual TLS (mTLS), Basic Authentication, or custom headers:

<Tabs>
<Tab name="mTLS">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      endpoint:
        value: https://backend.example.com/otlp:4317
      tls:
        cert:
          value: |
            -----BEGIN CERTIFICATE-----
            ...
        key:
          value: |
            -----BEGIN RSA PRIVATE KEY-----
            ...
```
</Tab>
<Tab name="Basic Authentication">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      endpoint:
        value: https://backend.example.com/otlp:4317
      authentication:
        basic:
          user:
            value: myUser
          password:
            value: myPwd
```
</Tab>
<Tab name="Token-based authentication with custom headers">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      endpoint:
        value: https://backend.example.com/otlp:4317
      headers:
      - name: Authorization
        prefix: Bearer
        value: "myToken"
```
</Tab>
</Tabs>

### 3b. Add Authentication Details From Secrets

Integrations into external systems usually need authentication details dealing with sensitive data. To handle that data properly in Secrets, TracePipeline supports the reference of Secrets.

Using the **valueFrom** attribute, you can map Secret keys for mutual TLS (mTLS), Basic Authentication, or with custom headers.

You can store the value of the token in the referenced Secret without any prefix or scheme, and you can configure it in the `headers` section of the TracePipeline. In the following example, the token has the prefix "Bearer".

<Tabs>
<Tab name="mTLS">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      endpoint:
        value: https://backend.example.com/otlp:4317
      tls:
        cert:
          valueFrom:
            secretKeyRef:
              name: backend
              namespace: default
              key: cert
        key:
          valueFrom:
            secretKeyRef:
              name: backend
              namespace: default
              key: key
```
</Tab>
<Tab name="Basic Authentication">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: backend
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
</Tab>
<Tab name="Token-based authentication with custom headers">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: TracePipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      endpoint:
        value: https://backend.example.com:4317
      headers:
      - name: Authorization
        prefix: Bearer
        valueFrom:
          secretKeyRef:
              name: backend
              namespace: default
              key: token
```
</Tab>
</Tabs>

The related Secret must have the referenced name, be located in the referenced namespace, and contain the mapped key. See the following example:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: backend
  namespace: default
stringData:
  endpoint: https://backend.example.com:4317
  user: myUser
  password: XXX
  token: YYY
```

### 4. Rotate the Secret

Telemetry Manager continuously watches the Secret referenced with the **secretKeyRef** construct. You can update the Secret’s values, and Telemetry Manager detects the changes and applies the new Secret to the setup.

> [!TIP]
> If you use a Secret owned by the [SAP BTP Service Operator](https://github.com/SAP/sap-btp-service-operator), you can configure an automated rotation using a `credentialsRotationPolicy` with a specific `rotationFrequency` and don’t have to intervene manually.

### 5. Deploy the Pipeline

To activate the TracePipeline, apply the `tracepipeline.yaml`  resource file in your cluster:

```bash
kubectl apply -f tracepipeline.yaml
```

### Result

You activated a TracePipeline and traces start streaming to your backend.

To check that the pipeline is running, wait until the status conditions of the TracePipeline in your cluster have status `True`:

```bash
kubectl get tracepipeline
NAME      CONFIGURATION GENERATED   GATEWAY HEALTHY   FLOW HEALTHY
backend   True                      True              True        
```

## Kyma Modules With Tracing Capabilities

Kyma bundles several modules that can be involved in user flows. Applications involved in a distributed trace must propagate the trace context to keep the trace complete. Optionally, they can enrich the trace with custom spans, which requires reporting them to the backend.

### Istio

The Istio module is crucial in distributed tracing because it provides the [Ingress Gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/). Typically, this is where external requests enter the cluster scope and are enriched with trace context if it hasn’t happened earlier. Furthermore, every component that’s part of the Istio Service Mesh runs an Istio proxy, which propagates the context properly but also creates span data. If Istio tracing is activated and taking care of trace propagation in your application, you get a complete picture of a trace, because every component automatically contributes span data. Also, Istio tracing is pre-configured to be based on the vendor-neutral [W3C Trace Context](https://www.w3.org/TR/trace-context/) protocol.

The Istio module is configured with an [extension provider](https://istio.io/latest/docs/tasks/observability/telemetry/) called `kyma-traces`. To activate the provider on the global mesh level using the Istio [Telemetry API](https://istio.io/latest/docs/reference/config/telemetry/#Tracing), place a resource to the `istio-system` namespace. The following code samples help setting up the Istio tracing feature:

<Tabs>
<Tab name="Extension Provider">

The following example configures all Istio proxies with the `kyma-traces` extension provider, which, by default, reports span data to the trace gateway of the Telemetry module.

```yaml
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: mesh-default
  namespace: istio-system
spec:
  tracing:
  - providers:
    - name: "kyma-traces"
```
</Tab>
<Tab name="Sampling Rate">

By default, the sampling rate is configured to 1%. That means that only 1 trace out of 100 traces is reported to the trace gateway, and all others are dropped. The sampling decision itself is propagated as part of the [trace context](https://www.w3.org/TR/trace-context/#sampled-flag) so that either all involved components are reporting the span data of a trace, or none.

> [!TIP]
> If you increase the sampling rate, you send more data your tracing backend and cause much higher network utilization in the cluster.
> To reduce costs and performance impacts in a production setup, a very low percentage of around 5% is recommended.

To configure an "always-on" sampling, set the sampling rate to 100%:

```yaml
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: mesh-default
  namespace: istio-system
spec:
  tracing:
  - providers:
    - name: "kyma-traces"
    randomSamplingPercentage: 100.00
```
</Tab>
<Tab name="Namespaces or Workloads">

If you need specific settings for individual namespaces or workloads, place additional Telemetry resources. If you don't want to report spans at all for a specific workload, activate the `disableSpanReporting` flag with the selector expression.

```yaml
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: tracing
  namespace: my-namespace
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: "my-app"
  tracing:
  - providers:
    - name: "kyma-traces"
    randomSamplingPercentage: 100.00
```
</Tab>
<Tab name="Trace Context Without Spans">

To enable the propagation of the [W3C Trace Context](https://www.w3.org/TR/trace-context/) only, without reporting any spans (so the actual tracing feature is disabled), you must enable the `kyma-traces` provider with a sampling rate of 0. With this configuration, you get the relevant trace context into the [access logs](https://kyma-project.io/#/istio/user/tutorials/01-45-enable-istio-access-logs) without any active trace reporting.

  ```yaml
  apiVersion: telemetry.istio.io/v1
  kind: Telemetry
  metadata:
    name: mesh-default
    namespace: istio-system
  spec:
    tracing:
    - providers:
      - name: "kyma-traces"
      randomSamplingPercentage: 0
  ```
</Tab>
</Tabs>

### Eventing

The [Eventing](https://kyma-project.io/#/eventing-manager/user/README) module uses the [CloudEvents](https://cloudevents.io/) protocol (which natively supports the [W3C Trace Context](https://www.w3.org/TR/trace-context) propagation). Because of that, it propagates trace context properly. However, it doesn't enrich a trace with more advanced span data.

### Serverless

By default, all engines for the [Serverless](https://kyma-project.io/#/serverless-manager/user/README) module integrate the [Open Telemetry SDK](https://opentelemetry.io/docs/reference/specification/metrics/sdk/). Thus, the used middlewares are configured to automatically propagate the trace context for chained calls.

Because the Telemetry endpoints are configured by default, Serverless also reports custom spans for incoming and outgoing requests. You can [customize Function traces](https://kyma-project.io/#/serverless-manager/user/tutorials/01-100-customize-function-traces) to add more spans as part of your Serverless source code.

## Operations

A TracePipeline runs several OTel Collector instances in your cluster. This Deployment serves OTLP endpoints and ships received data to the configured backend.

The Telemetry module ensures that the OTel Collector instances are operational and healthy at any time, for example, with buffering and retries. However, there may be situations when the instances drop traces, or cannot handle the trace load.

To detect and fix such situations, check the [pipeline status](./resources/04-tracepipeline.md#tracepipeline-status) and check out [Troubleshooting](#troubleshooting). If you have set up [pipeline health monitoring](./04-metrics.md#5-monitor-pipeline-health), check the alerts and reports in an integrated backend like [SAP Cloud Logging](./integration/sap-cloud-logging/README.md#use-sap-cloud-logging-alerts).

> [! WARNING]
> It's not recommended to access the metrics endpoint of the used OTel Collector instances directly, because the exposed metrics are no official API of the Kyma Telemetry module. Breaking changes can happen if the underlying OTel Collector version introduces such.
> Instead, use the [pipeline status](./resources/04-tracepipeline.md#tracepipeline-status).

## Limitations

- **Throughput**: Assuming an average span with 40 attributes with 64 characters, the maximum throughput is 4200 span/sec ~= 15.000.000 spans/hour. If this limit is exceeded, spans are refused. To increase the maximum throughput, manually scale out the gateway by increasing the number of replicas for the trace gateway. See [Module Configuration and Status](https://kyma-project.io/#/telemetry-manager/user/01-manager?id=module-configuration).
- **Unavailability of Output**: For up to 5 minutes, a retry for data is attempted when the destination is unavailable. After that, data is dropped.
- **No Guaranteed Delivery**: The used buffers are volatile. If the OTel Collector instance crashes, trace data can be lost.
- **Multiple TracePipeline Support**: The maximum amount of TracePipeline resources is 5.
- **System Span Filtering**: System-related spans reported by Istio are filtered out without the opt-out option, for example:
  - Any communication of applications to the Telemetry gateways
  - Any communication from the gateways to backends

## Troubleshooting

### No Spans Arrive at the Backend

**Symptom**: In the TracePipeline status, the `TelemetryFlowHealthy` condition has status **GatewayAllTelemetryDataDropped**.

**Cause**: Incorrect backend endpoint configuration (such as using the wrong authentication credentials), or the backend is unreachable.

**Solution**:

1. Check the `telemetry-trace-gateway` Pods for error logs by calling `kubectl logs -n kyma-system {POD_NAME}`.
2. Check if the backend is up and reachable.
3. Fix the errors.

### Not All Spans Arrive at the Backend

**Symptom**:

- The backend is reachable and the connection is properly configured, but some spans are refused.
- In the TracePipeline status, the `TelemetryFlowHealthy` condition has status **GatewaySomeTelemetryDataDropped**.

**Cause**: It can happen due to a variety of reasons - for example, the backend is limiting the ingestion rate.

**Solution**:

1. Check the `telemetry-trace-gateway` Pods for error logs by calling `kubectl logs -n kyma-system {POD_NAME}`. Also, check your observability backend to investigate potential causes.
2. If the backend is limiting the rate by refusing spans, try the options desribed in [Gateway Buffer Filling Up](#gateway-buffer-filling-up).
3. Otherwise, take the actions appropriate to the cause indicated in the logs.

### Custom Spans Don’t Arrive at the Backend, but Istio Spans Do

**Cause**: Your SDK version is incompatible with the OTel Collector version.

**Solution**:

1. Check which SDK version you are using for instrumentation.
2. Investigate whether it is compatible with the OTel Collector version.
3. If required, upgrade to a supported SDK version.

### Trace Backend Shows Fewer Traces than Expected

**Cause**: By [default](#istio), only 1% of the requests are sent to the trace backend for trace recording.

**Solution**:

To see more traces in the trace backend, increase the percentage of requests by changing the default settings.
If you just want to see traces for one particular request, you can manually force sampling:

1. Create a `values.yaml` file.
   The following example sets the value to `60`, which means 60% of the requests are sent to the tracing backend.

```yaml
  apiVersion: telemetry.istio.io/v1
  kind: Telemetry
  metadata:
    name: kyma-traces
    namespace: istio-system
  spec:
    tracing:
    - providers:
      - name: "kyma-traces"
      randomSamplingPercentage: 60
```

2. To override the default percentage, change the value for the **randomSamplingPercentage** attribute.
3. Deploy the `values.yaml` to your existing Kyma installation.

### Gateway Buffer Filling Up

**Symptom**: In the TracePipeline status, the `TelemetryFlowHealthy` condition has status **GatewayBufferFillingUp**.

**Cause**: The backend ingestion rate is too low compared to the gateway export rate.

**Solution**:

- Option 1: Increase the maximum backend ingestion rate - for example, by scaling out the SAP Cloud Logging instances.
- Option 2: Reduce the emitted spans in your applications.

### Gateway Throttling

**Symptom**:

- In the TracePipeline status, the `TelemetryFlowHealthy` condition has status **GatewayThrottling**.
- Also, your application might have error logs indicating a refusal for sending traces to the gateway.

**Cause**: Gateway cannot receive spans at the given rate.

**Solution**: Manually scale out the gateway by increasing the number of replicas for the trace gateway. See [Module Configuration and Status](https://kyma-project.io/#/telemetry-manager/user/01-manager?id=module-configuration).
