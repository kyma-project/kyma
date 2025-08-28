# Application Logs (Fluent Bit)

With application logs, you can debug an application and derive the internal state of an application. When logs are emitted with the correct severity level and context, they're essential for observing an application.

## Overview

The Telemetry module provides the [Fluent Bit](https://fluentbit.io/) log agent for the collection and shipment of application logs of any container running in the Kyma runtime.

You can configure the log agent with external systems using runtime configuration with a dedicated Kubernetes API ([CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions)) named `LogPipeline`. With the LogPipeline's HTTP output, you can natively integrate with vendors that support this output, or with any vendor using a [Fluentd integration](https://medium.com/hepsiburadatech/fluent-logging-architecture-fluent-bit-fluentd-elasticsearch-ca4a898e28aa).

The feature is optional, if you don't want to use the Logs feature, simply don't set up a LogPipeline.

<!--- custom output/unsupported mode is not part of Help Portal docs --->
If you want more flexibility than provided by the proprietary protocol, you can run the agent in the [unsupported mode](#unsupported-mode), using the full vendor-specific output options of Fluent Bit. If you need advanced configuration options, you can also bring your own log agent.

## Prerequisites

Your application must log to `stdout` or `stderr`, which ensures that the logs can be processed by Kubernetes primitives like `kubectl logs`. For details, see [Kubernetes: Logging Architecture](https://kubernetes.io/docs/concepts/cluster-administration/logging/).

## Architecture

In the Kyma cluster, the Telemetry module provides a DaemonSet of [Fluent Bit](https://fluentbit.io/) acting as a agent. The agent tails container logs from the Kubernetes container runtime and ships them to a backend.

![Architecture](./assets/logs-fluentbit-arch.drawio.svg)

1. Application containers print logs to `stdout/stderr` and are stored by the Kubernetes container runtime under the `var/log` directory and its subdirectories on the related Node.
2. Fluent Bit runs as a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) (one instance per Node), detects any new log files in the folder, and tails them using a filesystem buffer for reliability.
3. Fluent Bit discovers additional Pod metadata, such as Pod annotations and labels.
4. Telemetry Manager configures Fluent Bit with your output configuration, observes the log flow, and reports problems in the LogPipeline status.
5. The log agent sends the data to the observability system that's specified in your `LogPipeline` resource - either within the Kyma cluster, or, if authentication is set up, to an external observability backend. You can use the integration with HTTP to integrate a system directly or with an additional Fluentd installation.
6. To analyze and visualize your logs, access the internal or external observability system.

### Telemetry Manager

The LogPipeline resource is watched by Telemetry Manager, which is responsible for generating the custom parts of the Fluent Bit configuration.

![Manager resources](./assets/logs-fluentbit-resources.drawio.svg)

1. Telemetry Manager watches all LogPipeline resources and related Secrets.
2. Furthermore, Telemetry Manager takes care of the full lifecycle of the Fluent Bit DaemonSet itself. Only if you defined a LogPipeline, the agent is deployed.
3. Whenever the configuration changes, Telemetry Manager validates the configuration (with a [validating webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)) and generates a new configuration for the Fluent Bit DaemonSet, where several ConfigMaps for the different aspects of the configuration are generated.
4. Referenced Secrets are copied into one Secret that is also mounted to the DaemonSet.

### Log Agent

If a LogPipeline is defined, a DaemonSet is deployed acting as an agent. The agent is based on [Fluent Bit](https://fluentbit.io/) and encompasses the collection of application logs provided by the Kubernetes container runtime. The agent sends all data to the configured backend.

### Pipelines
<!--- Pipelines is not part of Help Portal docs --->
Fluent Bit comes with a pipeline concept, which supports a flexible combination of inputs with outputs and filtering in between. For details, see [Fluent Bit: Output](https://docs.fluentbit.io/manual/data-pipeline/outputs).
Kyma's Telemetry module brings a predefined setup of the Fluent Bit DaemonSet and a base configuration, which assures that the application logs of the workloads in the cluster are processed reliably and efficiently. Additionally, the Telemetry module provides a Kubernetes API called `LogPipeline` to configure outputs with some filtering capabilities.

This approach ensures reliable buffer management and isolation of pipelines, while keeping flexibility on customizations.

![Pipeline Concept](./assets/logs-fluentbit-pipelines.drawio.svg)

1. A dedicated `tail` **input** plugin reads the application logs, which are selected in the input section of the `LogPipeline`. Each `tail` input uses a dedicated `tag` with the name `<logpipeline>.*`.

2. The application logs are enriched by the `kubernetes` **filter**. You can add your own filters to the default filters.

3. Based on the default and custom filters, you get the desired **output** for each `LogPipeline`.

This approach assures a reliable buffer management and isolation of pipelines, while keeping flexibility on customizations.

## Setting up a LogPipeline

In the following steps, you can see how to construct and deploy a typical LogPipeline. Learn more about the available [parameters and attributes](resources/02-logpipeline.md).

### 1. Create a LogPipeline and Output

To ship application logs to a new output, create a resource of the kind `LogPipeline` and save the file (named, for example, `logpipeline.yaml`).

```yaml
kind: LogPipeline
apiVersion: telemetry.kyma-project.io/v1alpha1
metadata:
  name: http-backend
spec:
  output:
    http:
      dedot: false
      port: "80"
      uri: "/"
      host:
        value: https://myhost/logs
      user:
        value: "user"
      password:
        value: "not-required"
```

An output is a data destination configured by a [Fluent Bit output](https://docs.fluentbit.io/manual/pipeline/outputs) of the relevant type. The LogPipeline supports the following output types:

- **http**, which sends the data to the specified HTTP destination. The output is designed to integrate with a [Fluentd HTTP Input](https://docs.fluentd.org/input/http), which opens up a huge ecosystem of integration possibilities.
<!--- custom output/unsupported mode is not part of Help Portal docs --->
- **custom**, which supports the configuration of any destination in the Fluent Bit configuration syntax.

> [!WARNING]
> If you use a `custom` output, you put the LogPipeline in the [unsupported mode](#unsupported-mode).

See the following example of the `custom` output:

```yaml
spec:
  output:
    custom: |
      Name               http
      Host               https://myhost/logs
      Http_User          user
      Http_Passwd        not-required
      Format             json
      Port               80
      Uri                /
      Tls                on
      tls.verify         on
```

### 2. Filter Your Input

By default, input is collected from all namespaces, except the system namespaces (`kube-system`, `istio-system`, `kyma-system`), which are excluded by default.

To filter your application logs by namespace or container, use an input spec to restrict or specify which resources you want to include. For example, you can define the namespaces to include in the input collection, exclude namespaces from the input collection, or choose that only system namespaces are included. Learn more about the available [parameters and attributes](resources/02-logpipeline.md).

The following example collects input from all namespaces excluding `kyma-system` and only from the `istio-proxy` containers:

```yaml
kind: LogPipeline
apiVersion: telemetry.kyma-project.io/v1alpha1
metadata:
  name: http-backend
spec:
  input:
    application:
      namespaces:
        exclude:
          - kyma-system
      containers:
        include:
          - istio-proxy
  output:
    ...
```

<!--- custom filters/unsupported mode is not part of Help Portal docs --->

If filtering by namespace and container is not enough, use [Fluent Bit filters](https://docs.fluentbit.io/manual/data-pipeline/filters) to enrich logs for filtering by attribute, or to drop whole lines.

> [!WARNING]
> If you use a `custom` filter, you put the LogPipeline in the [unsupported mode](#unsupported-mode).

The following example uses the filter types [grep](https://docs.fluentbit.io/manual/pipeline/filters/grep) and [record_modifier](https://docs.fluentbit.io/manual/pipeline/filters/record-modifier), which are executed in sequence:

- The first filter keeps all log records that have the **kubernetes.labels.app** attribute set with the value `my-deployment`; all other logs are discarded. The **kubernetes** attribute is available for every log record. For more details, see [Kubernetes filter (metadata)](#kubernetes-filter-metadata).
- The second filter drops all log records fulfilling the given rule. In the example, typical namespaces are dropped based on the **kubernetes** attribute.
- Lastly, the record modifier adds a new attribute: Every log record is enriched with the cluster Node name as cluster identifier, for later filtering in the backend system. As the value, a placeholder refers to a Kubernetes-specific environment variable.

```yaml
kind: LogPipeline
apiVersion: telemetry.kyma-project.io/v1alpha1
metadata:
  name: http-backend
spec:
  filters:
    - custom: |
        Name    grep
        Regex   $kubernetes['labels']['app'] my-deployment
    - custom: |
        Name    grep
        Exclude $kubernetes['namespace_name'] kyma-system|kube-system|istio-system
    - custom: |
        Name    record_modifier
        Record  cluster_identifier ${KUBERNETES_SERVICE_HOST}
  input:
    ...
  output:
    ...
```

> [!WARNING]
> If you use a `custom` filter, you put the LogPipeline in the [unsupported mode](#unsupported-mode).

Telemetry Manager supports different types of [Fluent Bit filter](https://docs.fluentbit.io/manual/data-pipeline/filters). The example uses the filters [grep](https://docs.fluentbit.io/manual/pipeline/filters/grep) and [record_modifier](https://docs.fluentbit.io/manual/pipeline/filters/record-modifier).

- The first filter keeps all log records that have the **kubernetes.labels.app** attribute set with the value `my-deployment`; all other logs are discarded. The **kubernetes** attribute is available for every log record. For more details, see [Kubernetes filter (metadata)](#kubernetes-filter-metadata).
- The second filter drops all log records fulfilling the given rule. In the example, typical namespaces are dropped based on the **kubernetes** attribute.
- A log record is modified by adding a new attribute. In the example, a constant attribute is added to every log record to record the actual cluster Node name at the record for later filtering in the backend system. As a value, a placeholder is used referring to a Kubernetes-specific environment variable.

### 3. Add Authentication Details From Secrets

Integrations into external systems usually need authentication details dealing with sensitive data. To handle that data properly in Secrets, LogPipeline supports the reference of Secrets.

Using the **http** output definition and the **valueFrom** attribute, you can map Secret keys for mutual TLS (mTLS) or Basic Authentication:

<!-- tabs:start -->

#### mTLS

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: http-backend
spec:
  output:
    http:
      dedot: false
      port: "80"
      uri: "/"
      host:
        valueFrom:
            secretKeyRef:
              name: http-backend-credentials
              namespace: default
              key: HTTP_ENDPOINT
      tls:
        cert:
          valueFrom:
            secretKeyRef:
              name: http-backend-credentials
              namespace: default
              key: TLS_CERT
        key:
          valueFrom:
            secretKeyRef:
              name: http-backend-credentials
              namespace: default
              key: TLS_KEY
  input:
    ...
  filters:
    ...
```

#### Basic Authentication

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: http-backend
spec:
  output:
    http:
      dedot: false
      port: "80"
      uri: "/"
      host:
        valueFrom:
          secretKeyRef:
            name: http-backend-credentials
            namespace: default
            key: HTTP_ENDPOINT
      user:
        valueFrom:
          secretKeyRef:
            name: http-backend-credentials
            namespace: default
            key: HTTP_USER
      password:
        valueFrom:
          secretKeyRef:
            name: http-backend-credentials
            namespace: default
            key: HTTP_PASSWORD
  input:
    ...
  filters:
    ...
```

<!-- tabs:end -->

The related Secret must have the referenced name, be located in the referenced namespace, and contain the mapped key. See the following example:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: http-backend-credentials
stringData:
  HTTP_ENDPOINT: https://myhost/logs
  HTTP_USER: myUser
  HTTP_PASSWORD: XXX
  TLS_CERT: ...
  TLS_KEY: ...
```

<!--- custom output/unsupported mode is not part of Help Portal docs --->
To use data provided by the Kubernetes Secrets in a `custom` output definition, use placeholder expressions for the data provided by the Secret, then specify the actual mapping to the Secret keys in the **variables** section, like in the following example:

```yaml
kind: LogPipeline
apiVersion: telemetry.kyma-project.io/v1alpha1
metadata:
  name: http-backend
spec:
  output:
    custom: |
      Name               http
      Host               ${ENDPOINT} # Defined in Secret
      HTTP_User          ${USER} # Defined in Secret
      HTTP_Password      ${PASSWORD} # Defined in Secret
      Tls                On
  variables:
    - name: ENDPOINT
      valueFrom:
        secretKeyRef:
        - name: http-backend-credentials
          namespace: default
          key: HTTP_ENDPOINT
  input:
    ...
  filters:
    ...
```

> [!WARNING]
> If you use a `custom` output, you put the LogPipeline in the [unsupported mode](#unsupported-mode).

### 4. Rotate the Secret

Telemetry Manager continuously watches the Secret referenced with the **secretKeyRef** construct. You can update the Secret’s values, and Telemetry Manager detects the changes and applies the new Secret to the setup.

> [!TIP]
> If you use a Secret owned by the [SAP BTP Service Operator](https://github.com/SAP/sap-btp-service-operator), you can configure an automated rotation using a `credentialsRotationPolicy` with a specific `rotationFrequency` and don’t have to intervene manually.

### 5. Deploy the Pipeline

To activate the LogPipeline, apply the  `logpipeline.yaml` resource file in your cluster:

```bash
kubectl apply -f logpipeline.yaml
```

### Result

You activated a LogPipeline and logs start streaming to your backend.

To check that the pipeline is running, wait until the status conditions of the LogPipeline in your cluster have status `True`:

```bash
kubectl get logpipeline
NAME      CONFIGURATION GENERATED   AGENT HEALTHY   FLOW HEALTHY
backend   True                      True            True        
```

## Log Record Processing

After a log record has been read, it is preprocessed by configured plugins, like the `kubernetes` filter. Thus, when a record is ready to be processed by the sections defined in the LogPipeline definition, it has several attributes available for processing and shipment.

![Flow](./assets/logs-fluentbit-flow.drawio.svg)

Learn more about the flow of the log record through the general pipeline and the available log attributes in the following stages:

### Container Log Message

The following example assumes that there’s a container `myContainer` of Pod `myPod`, running in namespace `myNamespace`, logging to `stdout` with the following log message in the JSON format:

```json
{
  "level": "warn",
  "message": "This is the actual message",
  "tenant": "myTenant",
  "traceID": "123"
}
```

### Tail Input

The `tail` input plugin reads the log message from a log file managed by the container runtime. The input plugin brings a dedicated filesystem buffer for the pipeline. The file name contains the namespace, Pod, and container information that will be available later as part of the [tag](https://docs.fluentbit.io/manual/concepts/key-concepts#tag). The tag is prefixed with the pipeline name. The resulting log record available in an internal Fluent Bit representation looks similar to the following example:

```json
{
  "time": "2022-05-23T15:04:52.193317532Z",
  "stream": "stdout",
  "_p": "F",
  "log": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"traceID\": \"123\"}"
}
```

The attributes have the following meaning:

- **time**: The timestamp generated by the container runtime at the moment the log was written to the log file.
- **stream**: The stream to which the application wrote the log, either `stdout` or `stderr`.
- **_p**: Indicates if the log message is partial (`P`) or final (`F`). Optional, dependent on container runtime. Because a CRI multiline parser is applied for the tailing phase, all multilines on the container runtime level are aggregated already and no partial entries must be left.
- **log**: The raw and unparsed log message.

### Kubernetes Filter (Metadata)

After the tail input, the [Kubernetes filter](https://docs.fluentbit.io/manual/pipeline/filters/kubernetes) is applied. The container information from the log file name (available in the tag) is interpreted and used for a Kubernetes API Server request to resolve more metadata of the container. All the resolved metadata enrich the existing record as a new attribute `kubernetes`:

```json
{
  "kubernetes":
  {
      "pod_name": "myPod-74db47d99-ppnsw",
      "namespace_name": "myNamespace",
      "pod_id": "88dbd1ef-d977-4636-804d-ef220454be1c",
      "host": "myHost1",
      "container_name": "myContainer",
      "docker_id": "5649c36fcc1e956fc95e3145441f427d05d6e514fa439f4e4f1ccee80fb2c037",
      "container_hash": "myImage@sha256:1f8d852989c16345d0e81a7bb49da231ade6b99d51b95c56702d04c417549b26",
      "container_image": "myImage:myImageTag",
      "labels":
      {
          "app": "myApp",
          "sidecar.istio.io/inject": "true",
          ...
      }
  }
}
```

### Kubernetes Filter (JSON Parser)

After the enrichment of the log record with the Kubernetes-relevant metadata, the [Kubernetes filter](https://docs.fluentbit.io/manual/pipeline/filters/kubernetes) also tries to parse the record as a JSON document. If that is successful, all the parsed root attributes of the parsed document are added as new individual root attributes of the log.

The record **before** applying the JSON parser:

```json
{
  "time": "2022-05-23T15:04:52.193317532Z",
  "stream": "stdout",
  "_p": "F",
  "log": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"traceID\": \"123\"}",
  "kubernetes": {...}
}
```

The record **after** applying the JSON parser:

```json
{
  "time": "2022-05-23T15:04:52.193317532Z",
  "stream": "stdout",
  "_p": "F",
  "log": "{\"level\": \"warn\",\"message\": \"This is the actual message\",\"tenant\": \"myTenant\",\"traceID\": \"123\"}",
  "kubernetes": {...},
  "level": "warn",
  "message": "This is the actual message",
  "tenant": "myTenant",
  "traceID": "123"
}
```

### Further Enrichment

Additionally, the agent enriches every log record with the `cluster_identifier` attribute by setting the APIServer URL of the underlying Kubernetes cluster:

```json
{
   "cluster_identifier": "<APIServer URL>"
   ...
}
```

For LogPipelines that use an HTTP output, the following attributes are enriched for optimized integration with SAP Cloud Logging:

```json
{
  "@timestamp": "<value of attribute 'time'>",
  "date": "<agent time in iso8601>",
  "kubernetes": {
    "app_name": "<value of Pod label 'app.kubernetes.io/name' or 'app'>"
    ...
  },
  ...
}
```

The enriched timestamp attributes have the following meaning:

- **time**: The time when the container runtime captured the log on `stdout/stderr`, which is very close to the time when the log originated in the application.
- **date**: The time when the log agent processed the log, which is later than the value in `time`.
- **@timestamp**: Contains the same value as **time**, optimized for SAP Cloud Logging integration.

## Operations

The Telemetry module ensures that the log agent instances are operational and healthy at any time, for example, with buffering and retries. However, there may be situations when the instances drop logs, or cannot handle the log load.

To detect and fix such situations, check the [pipeline status](./resources/02-logpipeline.md#logpipeline-status) and check out [Troubleshooting](#troubleshooting). If you have set up [pipeline health monitoring](./04-metrics.md#5-monitor-pipeline-health), check the alerts and reports in an integrated backend like [SAP Cloud Logging](./integration/sap-cloud-logging/README.md#use-sap-cloud-logging-alerts).

> [!WARNING]
> It's not recommended to access the metrics endpoint of the used FluentBit instances directly, because the exposed metrics are no official API of the Kyma Telemetry module. Breaking changes can happen if the underlying FluentBit version introduces such.
> Instead, use the [pipeline status](./resources/02-logpipeline.md#logpipeline-status).

## Limitations

- **Reserved Log Attributes**: The log attribute named `kubernetes` is a special attribute that’s enriched by the `kubernetes` filter. When you use that attribute as part of your structured log payload, the metadata enriched by the filter are overwritten by the payload data. Filters that rely on the original metadata might no longer work as expected.
- **Buffer Limits**: Fluent Bit buffers up to 1 GB of logs if a configured output cannot receive logs. The oldest logs are dropped when the limit is reached or after 300 retries.
- **Throughput**: Each Fluent Bit Pod (each running on a dedicated Node) can process up to 10 MB/s of logs for a single LogPipeline. With multiple pipelines, the throughput per pipeline is reduced. The used logging backend or performance characteristics of the output plugin might limit the throughput earlier.
- **Multiple LogPipeline Support**: The maximum amount of LogPipeline resources is 5.

### Unsupported Mode
<!--- unsupported mode is not part of Help Portal docs --->
The `unsupportedMode` attribute of a LogPipeline indicates that you are using a `custom` filter and/or `custom` output. The Kyma team does not provide support for a custom configuration.

### Fluent Bit Plugins
<!--- Fluent Bit Plugins is not part of Help Portal docs --->

You cannot enable the following plugins, because they potentially harm the stability:

- Kubernetes Filter
- Rewrite_Tag Filter

## Troubleshooting

### No Logs Arrive at the Backend

**Symptom**:

- No logs arrive at the backend.
- In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **AgentAllTelemetryDataDropped**.

**Cause**: Incorrect backend endpoint configuration (for example, using the wrong authentication credentials) or the backend being unreachable.

**Solution**:

- Check the `telemetry-fluent-bit` Pods for error logs by calling `kubectl logs -n kyma-system {POD_NAME}`.
- Check if the backend is up and reachable.

### Not All Logs Arrive at the Backend

**Symptom**:

- The backend is reachable and the connection is properly configured, but some logs are refused.
- In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **AgentSomeTelemetryDataDropped**.

**Cause**: It can happen due to a variety of reasons. For example, a possible reason may be that the backend is limiting the ingestion rate, or the backend is refusing logs because they are too large.

**Solution**:

1. Check the `telemetry-fluent-bit` Pods for error logs by calling `kubectl logs -n kyma-system {POD_NAME}`. Also, check your observability backend to investigate potential causes.
2. If the backend is limiting the rate by refusing logs, try the options described in [Agent Buffer Filling Up](#agent-buffer-filling-up).
3. Otherwise, take the actions appropriate to the cause indicated in the logs.

### Agent Buffer Filling Up

**Symptom**: In the LogPipeline status, the `TelemetryFlowHealthy` condition has status **AgentBufferFillingUp**.

**Cause**: The backend ingestion rate is too low compared to the log collection rate.

**Solution**:

- Option 1: Increase maximum backend ingestion rate. For example, by scaling out the SAP Cloud Logging instances.

- Option 2: Reduce emitted logs by re-configuring the LogPipeline (for example, by applying namespace or container filters).

- Option 3: Reduce emitted logs in your applications (for example, by changing severity level).
