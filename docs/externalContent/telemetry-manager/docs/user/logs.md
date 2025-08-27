# Application Logs (OTLP)

With application logs, you can debug an application and derive the internal state of an application. The Telemetry module supports observing your applications with logs of the correct severity level and context.

## Overview

The Telemetry module provides a log gateway for push-based collection of logs using OTLP and, optionally, an agent for the collection of logs of any container printing logs to the `stdout/stderr` channel running in the Kyma runtime. Kyma modules like [Istio](https://kyma-project.io/#/istio/user/README) contribute access logs. The Telemetry module enriches the data and ships them to your chosen backend (see [Vendors who natively support OpenTelemetry](https://opentelemetry.io/ecosystem/vendors/)).

You can configure the log gateway and agent with external systems using runtime configuration with a dedicated Kubernetes API ([CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions)) named LogPipeline.
The Log feature is optional. If you don't want to use it, simply don't set up a LogPipeline.

## Prerequisites

- Before you can collect logs from a component, it must emit the logs. Typically, it uses a logger framework for the used language runtime (like Node.js) and prints them to the `stdout` or `stderr` channel ([Kubernetes: How nodes handle container logs](https://kubernetes.io/docs/concepts/cluster-administration/logging/#how-nodes-handle-container-logs)). Alternatively, you can use the [OTel SDK](https://opentelemetry.io/docs/languages/) to use the [push-based OTLP format](https://opentelemetry.io/docs/specs/otlp/).

- If you want to emit the logs to the `stdout/stderr` channel, use structured logs in a JSON format with a logger library like log4J. With that, the log agent can parse your log and enrich all JSON attributes as log attributes, and a backend can use that.

- If you prefer the push-based alternative with OTLP, also use a logger library like log4J. However, you additionally instrument that logger and bridge it to the OTel SDK. For details, see [OpenTelemetry: New First-Party Application Logs](https://opentelemetry.io/docs/specs/otel/logs/#new-first-party-application-logs).

## Architecture

In the Telemetry module, a central in-cluster Deployment of an [OTel Collector](https://opentelemetry.io/docs/collector/) acts as a gateway. The gateway exposes endpoints for the [OpenTelemetry Protocol (OTLP)](https://opentelemetry.io/docs/specs/otlp/) for GRPC and HTTP-based communication using the dedicated `telemetry-otlp-logs` service, to which your applications send the logs data.

You can choose whether you also want an agent, based on a DaemonSet of an OTel Collector. This agent can tail logs of a container from the underlying container runtime.

![Architecture](./assets/logs-arch.drawio.svg)

1. Application containers print JSON logs to the `stdout/stderr` channel and are stored by the Kubernetes container runtime under the `var/log` directory and its subdirectories at the related Node. Istio is configured to write access logs to `stdout` as well.
2. If you choose to use the agent, an OTel Collector runs as a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) (one instance per Node), detects any new log files in the folder, and tails and parses them.
3. An application (exposing logs in OTLP) sends logs to the central log gateway service. Istio is configured to push access logs with OTLP as well.
4. The gateway and agent discover the metadata and enrich all received data with metadata of the source by communicating with the Kubernetes APIServer. Furthermore, they filter data according to the pipeline configuration.
5. Telemetry Manager configures the agent and gateway according to the `LogPipeline` resource specification, including the target backend. Also, it observes the logs flow to the backend and reports problems in the LogPipeline status.
8. The log agent and gateway send the data to the observability system that's specified in your `LogPipeline` resource - either within the Kyma cluster, or, if authentication is set up, to an external observability backend.
9. You can analyze the logs data with your preferred backend.

### Telemetry Manager

The LogPipeline resource is watched by Telemetry Manager, which is responsible for generating the custom parts of the OTel Collector configuration.

![Manager resources](./assets/logs-resources.drawio.svg)

1. Telemetry Manager watches all LogPipeline resources and related Secrets.
2. Furthermore, Telemetry Manager takes care of the full lifecycle of the gateway Deployment and the agent DaemonSet. Only if you defined a LogPipeline, the gateway and agent are deployed.
3. Whenever the user configuration changes, Telemetry Manager validates it and generates a single configuration for the gateway and agent.
4. Referenced Secrets are copied into one Secret that is mounted to the gateway as well.

### Log Gateway

In a Kyma cluster, the log gateway is the central component to which all components can send their individual logs. The gateway collects, enriches, and dispatches the data to the configured backend. For more information, see [Telemetry Gateways](./gateways.md).

### Log Agent

If you configure a feature in the `input` section of your LogPipeline, an additional DaemonSet is deployed acting as an agent. The agent is based on an [OTel Collector](https://opentelemetry.io/docs/collector/) and encompasses the collection and conversion of logs from the container runtime. Hereby, the workload container just prints the structured log to the `stdout/stderr` channel. The agent picks them up, parses and enriches them, and sends all data in OTLP to the configured backend.

## Setting up a LogPipeline

In the following steps, you can see how to construct and deploy a typical LogPipeline. Learn more about the available [parameters and attributes](resources/02-logpipeline.md).

### 1. Create a LogPipeline

To ship logs to a new OTLP output, create a resource of the kind `Logipeline` and save the file (named, for example, `logpipeline.yaml`).

This configures the underlying OTel Collector with a pipeline for logs and opens a push endpoint that is accessible with the `telemetry-otlp-logs` service. For details, see [Gateway Usage](./gateways.md#usage). The following push URLs are set up:

- GRPC: `http://telemetry-otlp-logs.kyma-system:4317`
- HTTP: `http://telemetry-otlp-logs.kyma-system:4318`

The default protocol for shipping the data to a backend is GRPC, but you can choose HTTP instead. Depending on the configured protocol, an `otlp` or an `otlphttp` exporter is used. Ensure that the correct port is configured as part of the endpoint.

<Tabs>
<Tab name="GRPC">

For GRPC, use:

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      endpoint:
        value: https://backend.example.com:4317
```
</Tab>
<Tab name="HTTP">

For HTTP, use the **protocol** attribute:

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: backend
spec:
  output:
    otlp:
      protocol: http
      endpoint:
        value: https://backend.example.com:4318
```
</Tab>
</Tabs>

### 2a. Add Authentication Details From Plain Text

To integrate with external systems, you must configure authentication details. You can use mutual TLS (mTLS), Basic Authentication, or custom headers:

<Tabs>
<Tab name="mTLS">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
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
kind: LogPipeline
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
kind: LogPipeline
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
### 2b. Add Authentication Details From Secrets

Integrations into external systems usually need authentication details dealing with sensitive data. To handle that data properly in Secrets, LogPipeline supports the reference of Secrets.

Using the **valueFrom** attribute, you can map Secret keys for mutual TLS (mTLS), Basic Authentication, or with custom headers.

You can store the value of the token in the referenced Secret without any prefix or scheme, and you can configure it in the headers section of the LogPipeline. In the following example, the token has the prefix “Bearer”.

<Tabs>
<Tab name="mTLS">

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
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
kind: LogPipeline
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
kind: LogPipeline
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

### 3. Rotate the Secret

Telemetry Manager continuously watches the Secret referenced with the **secretKeyRef** construct. You can update the Secret’s values, and Telemetry Manager detects the changes and applies the new Secret to the setup.

> [!TIP]
> If you use a Secret owned by the [SAP BTP Service Operator](https://github.com/SAP/sap-btp-service-operator), you can configure an automated rotation using a `credentialsRotationPolicy` with a specific `rotationFrequency` and don’t have to intervene manually.

### 4. Activate Application Input

To enable collection of logs printed by containers to the `stdout/stderr` channel, define a LogPipeline that has the `application` section enabled as input:

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: backend
spec:
  input:
    application:
      enabled: true
  output:
    otlp:
      ...
```

By default, input is collected from all namespaces, except the system namespaces `kube-system`, `istio-system`, `kyma-system`, which are excluded by default.

To filter your application logs by namespace or container, use an input spec to restrict or specify which resources you want to include. For example, you can define the namespaces to include in the input collection, exclude namespaces from the input collection, or choose that only system namespaces are included. Learn more about the available [parameters and attributes](resources/02-logpipeline.md).

The following pipeline collects input from all namespaces excluding `kyma-system` and only from the `istio-proxy` containers:

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: backend
spec:
  input:
    application:
      enabled: true
      namespaces:
        exclude:
          - myNamespace
      containers:
        exclude:
          - myContainer
    otlp:
      ...
```

After tailing the log files from the container runtime, the payload of the log lines is transformed into an OTLP entry. Learn more about the flow of the log record through the steps and the available log attributes in the following stages:

- [Log Tailing](#log-tailing)
- [JSON Parsing](#json-parsing)
- [Severity Parsing](#severity-parsing)
- [Trace Parsing](#trace-parsing)
- [Log Body Determination](#log-body-determination)

The following example assumes that there’s a container `myContainer` of Pod `myPod`, running in namespace `myNamespace`, logging to `stdout` with the following log message in the JSON format:

```json
{
  "level": "warn",
  "message": "This is the actual message",
  "tenant": "myTenant",
  "traceID": "123"
}
```

#### Log Tailing

The agent reads the log message from a log file managed by the container runtime. The file name contains namespace, Pod and Container information that will be available later as log attributes. The raw log record looks similar to the following example:

```json
{
  "time": "2022-05-23T15:04:52.193317532Z",
  "stream": "stdout",
  "_p": "F",
  "log": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"trace_id\": \"123\"}"
}
```

After the tailing, the created OTLP record looks like the following example:

```json
{
  "time": "2022-05-23T15:04:52.100000000Z",
  "observedTime": "2022-05-23T15:04:52.200000000",
  "attributes": {
   "log.file.path": "/var/log/pods/myNamespace_myPod-<containerID>/myContainer/<containerRestarts>.log",
   "log.iostream": "stdout"
  },
  "resourceAttributes": {
    "k8s.container.name": "myContainer",
    "k8s.container.restart_count": "<containerRestarts>",
    "k8s.pod.name": "myPod",
    "k8s.namespace.name": "myNamespace"
  },
  "body": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"trace_id\": \"123\"}"
}
```

All information identifying the source of the log (like the Container, Pod and namespace name) are enriched as resource attributes following the [Kubernetes conventions](https://opentelemetry.io/docs/specs/semconv/resource/k8s/). Further metadata - like the original file name and channel - are enriched as log attributes following the [log attribute conventions](https://opentelemetry.io/docs/specs/semconv/general/logs/). The **time** value provided in the container runtime log entry is used as **time** attribute in the new OTel record, as it is very close to the actual time when the log happened. Additionally, the **observedTime** is set with the time when the agent actual read the log record as recommended by the [OTel log specification](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-observedtimestamp). The log payload is moved to the OTLP **body** field.

#### JSON Parsing

If the value of the **body** is a JSON document, the value is parsed and all JSON root attributes are enriched as additional log attributes. The original body is moved into the **log.original** attribute (managed with the LogPipeline attribute **input.application.keepOriginalBody**: `true`).

The resulting OTLP record looks like the following example:

```json
{
  "time": "2022-05-23T15:04:52.100000000Z",
  "observedTime": "2022-05-23T15:04:52.200000000",
  "attributes": {
   "log.file.path": "/var/log/pods/myNamespace_myPod-<containerID>/myContainer/<containerRestarts>.log",
   "log.iostream": "stdout",
   "log.original": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"trace_id\": \"123\"}",
   "level": "warn",
   "tenant": "myTenant",
   "trace_id": "123",
   "message": "This is the actual message"
  },
  "resourceAttributes": {
    "k8s.container.name": "myContainer",
    "k8s.container.restart_count": "<containerRestarts>",
    "k8s.pod.name": "myPod",
    "k8s.namespace.name": "myNamespace"
  },
  "body": ""
}
```

#### Severity Parsing

Typically, a log message has a log level written to a field `level`. Based on that, the agent tries to parse the log attribute **level** with a [severity parser](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/operators/severity_parser.md). If that is successful, the log attribute is transformed into the OTel attributes **severityText** and **severityNumber**.

#### Trace Parsing

OTLP natively supports attaching trace context to log records. If possible, the log agent parses the following log attributes according to the [W3C-Tracecontext specification](https://www.w3.org/TR/trace-context/#traceparent-header):

- **trace_id**
- **span_id**
- **trace_flags**
- **traceparent**

#### Log Body Determination

Because the actual log message is typically located in the **body** attribute, the agent moves a log attribute called **message** (or **msg**) into the **body**.

At this point, before further enrichment, the resulting overall log record looks like the following example:

```json
{
  "time": "2022-05-23T15:04:52.100000000Z",
  "observedTime": "2022-05-23T15:04:52.200000000",
  "attributes": {
   "log.file.path": "/var/log/pods/myNamespace_myPod-<containerID>/myContainer/<containerRestarts>.log",
   "log.iostream": "stdout",
   "log.original": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"trace_id\": \"123\"}",
   "tenant": "myTenant",
  },
  "resourceAttributes": {
    "k8s.container.name": "myContainer",
    "k8s.container.restart_count": "<containerRestarts>",
    "k8s.pod.name": "myPod",
    "k8s.namespace.name": "myNamespace"
  },
  "body": "This is the actual message",
  "severityNumber": 13,
  "severityTex": "warn",
  "trace_id": 123
}
```

### 5. Deactivate OTLP Logs

If you have more than one backend, you can specify from which `input` logs are pushed to each backend. For example, if OTLP logs should go to one backend and only logs from the tail input to the other backend, then disable the OTLP input for the second backend.

By default, `otlp` input is enabled.

To drop the push-based OTLP logs that are received by the log gateway, define a LogPipeline that has the `otlp` section disabled as an input:

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: backend
spec:
  input:
    application:
      enabled: true
    otlp:
      disabled: true
  output:
    otlp:
      endpoint:
        value: https://backend.example.com:4317
```

With this, the agent starts collecting all container logs, while the push-based OTLP logs are dropped by the gateway.

### 6. Deploy the Pipeline

To activate the LogPipeline, apply the `logpipeline.yaml` resource file in your cluster:

```bash
kubectl apply -f logpipeline.yaml
```

### Result

You activated a LogPipeline and logs start streaming to your backend.

To check that the pipeline is running, wait until the status conditions of the LogPipeline in your cluster have status `True`:

```bash
kubectl get logpipeline
NAME      CONFIGURATION GENERATED   GATEWAY HEALTHY   AGENT HEALTHY   FLOW HEALTHY
backend   True                      True              True            True        
```

## Kyma Modules With Logging Capabilities

Kyma bundles modules that can be involved in user flows. If you want to collect all logs of all modules, enable the `application` input for the `kyma-system` namespace.

### Istio

The Istio module is crucial as it provides the [Ingress Gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/). Typically, this is where external requests enter the cluster scope. Furthermore, every component that’s part of the Istio Service Mesh runs an Istio proxy. Using the Istio telemetry API, you can enable access logs for the Ingress Gateway and the proxies individually.

The Istio module is configured with an [extension provider](https://istio.io/latest/docs/tasks/observability/telemetry/) called `kyma-logs`. To activate the provider on the global mesh level using the Istio [Telemetry API](https://istio.io/latest/docs/reference/config/telemetry), place a resource to the `istio-system` namespace.

The following example configures all Istio proxies with the `kyma-logs` extension provider, which, by default, reports access logs to the log gateway of the Telemetry module.

```yaml
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: mesh-default
  namespace: istio-system
spec:
  accessLogging:
    - providers:
        - name: kyma-logs
```

## Operations

A LogPipeline runs several OTel Collector instances in your cluster. This Deployment serves OTLP endpoints and ships received data to the configured backend.

The Telemetry module ensures that the OTel Collector instances are operational and healthy at any time, for example, with buffering and retries. However, there may be situations when the instances drop logs, or cannot handle the log load.

To detect and fix such situations, check the [pipeline status](./resources/02-logpipeline.md#logpipeline-status) and check out [Troubleshooting](#troubleshooting). If you have set up [pipeline health monitoring](./04-metrics.md#5-monitor-pipeline-health), check the alerts and reports in an integrated backend like [SAP Cloud Logging](./integration/sap-cloud-logging/README.md#use-sap-cloud-logging-alerts).

> [! WARNING]
> It's not recommended to access the metrics endpoint of the used OTel Collector instances directly, because the exposed metrics are no official API of the Telemetry module. Breaking changes can happen if the underlying OTel Collector version introduces such.
> Instead, use the [pipeline status](./resources/02-logpipeline.md#logpipeline-status).

## Limitations

- **Throughput**:
  - When pushing OTLP logs of an average size of 2KB to the log gateway, using its default configuration (two instances), the Telemetry module can process approximately 12,000 logs per second (LPS). To ensure availability, the log gateway runs with multiple instances. For higher throughput, manually scale out the gateway by increasing the number of replicas. See [Module Configuration and Status](https://kyma-project.io/#/telemetry-manager/user/01-manager?id=module-configuration). Ensure that the chosen scaling factor does not exceed the maximum throughput of the backend, as it may refuse logs if the rate is too high.
  - For example, to scale out the gateway for scenarios like a `Large` instance of SAP Cloud Logging (up to 30,000 LPS), you can raise the throughput to about 20,000 LPS by increasing the number of replicas to 4 instances.
  - The log agent, running one instance per node, handles tailing logs from stdout using the `runtime` input. When writing logs of an average size of 2KB to stdout, a single log agent instance can process approximately 9,000 LPS.
- **Load Balancing With Istio**: By design, the connections to the gateway are long-living connections (because OTLP is based on gRPC and HTTP/2). For optimal scaling of the gateway, the clients or applications must balance the connections across the available instances, which is automatically achieved if you use an Istio sidecar. If your application has no Istio sidecar, the data is always sent to one instance of the gateway.
- **Unavailability of Output**: For up to 5 minutes, a retry for data is attempted when the destination is unavailable. After that, data is dropped.
- **No Guaranteed Delivery**: The used buffers are volatile. If the gateway or agent instances crash, logs data can be lost.
- **Multiple LogPipeline Support**: The maximum amount of LogPipeline resources is 5.

## Troubleshooting

### No Logs Arrive at the Backend

**Symptom**:

- No logs arrive at the backend.
- In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **GatewayAllTelemetryDataDropped** or **AgentAllTelemetryDataDropped**.

**Cause**: Incorrect backend endpoint configuration (such as using the wrong authentication credentials) or the backend is unreachable.

**Solution**:

1. Check the error logs for the affected Pod by calling `kubectl logs -n kyma-system {POD_NAME}`:
   - For **GatewayAllTelemetryDataDropped**, check Pod `telemetry-log-gateway`.
   - For **AgentAllTelemetryDataDropped**, check Pod `telemetry-log-agent`.
2. Check if the backend is up and reachable.
3. Fix the errors.

### Not All Logs Arrive at the Backend

**Symptom**:

- The backend is reachable and the connection is properly configured, but some logs are refused.
- In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **GatewaySomeTelemetryDataDropped** or **AgentSomeTelemetryDataDropped**.

**Cause**: It can happen due to a variety of reasons - for example, the backend is limiting the ingestion rate.

**Solution**:

1. Check the error logs for the affected Pod by calling `kubectl logs -n kyma-system {POD_NAME}`:
   - For **GatewaySomeTelemetryDataDropped**, check Pod `telemetry-log-gateway`.
   - For **AgentSomeTelemetryDataDropped**, check Pod `telemetry-log-agent`.
2. Check your observability backend to investigate potential causes.
3. If the backend is limiting the rate by refusing logs, try the options described in [Buffer Filling Up](#buffer-filling-up).
4. Otherwise, take the actions appropriate to the cause indicated in the logs.

### No Access Logs With Push-Based OTLP LogPipeline

**Symptom**: The Istio telemetry resource uses a push-based OTLP LogPipeline. Even though this pipeline is healthy, access logs do not arrive at the backend.

**Cause**: Istio cannot discover the OTLP endpoint.

**Solution**: Recreate the Istio telemetry resource. For details, see [Istio](#istio).

### Buffer Filling Up

**Symptom**: In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **GatewayBufferFillingUp** or **AgentBufferFillingUp**.

**Cause**: The backend ingestion rate is too low compared to the export rate of the gateway or agent.

**Solution**:

- Option 1: Increase maximum backend ingestion rate. For example, by scaling out the SAP Cloud Logging instances.

- Option 2: Reduce emitted logs by re-configuring the LogPipeline (for example, by disabling certain inputs or applying namespace filters).

- Option 3: Reduce emitted logs in your applications.

### Gateway Throttling

**Symptom**: In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **GatewayThrottling**.

**Cause**: Gateway cannot receive logs at the given rate.

**Solution**: Manually scale out the gateway by increasing the number of replicas for the log gateway. See [Module Configuration and Status](https://kyma-project.io/#/telemetry-manager/user/01-manager?id=module-configuration).
