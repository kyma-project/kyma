---
title: Telemetry - LogPipeline
---

The `logpipeline.telemetry.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to filter and ship application logs in Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd logpipeline.telemetry.kyma-project.io -o yaml
```

## Sample custom resource

The following LogPipeline object defines a pipeline integrating with the HTTP/JSON-based output using basic authentication, excluding application logs emitted by istio-proxy containers.

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogPipeline
metadata:
  name: custom-fluentd
spec:
  input:
    application:
      containers:
        exclude:
        - istio-proxy
      namespaces: {}
  output:
    http:
      dedot: true
      host:
        valueFrom:
          secretKeyRef:
            key: Fluentd-endpoint
            name: custom-fluentd
            namespace: default
      password:
        valueFrom:
          secretKeyRef:
            key: Fluentd-password
            name: custom-fluentd
            namespace: default
      tls: {}
      uri: /customindex/kyma
      user:
        valueFrom:
          secretKeyRef:
            key: Fluentd-username
            name: custom-fluentd
            namespace: default
status:
  conditions:
  - lastTransitionTime: "2022-11-25T12:38:36Z"
    reason: FluentBitDaemonSetRestarted
    type: Pending
  - lastTransitionTime: "2022-11-25T12:39:26Z"
    reason: FluentBitDaemonSetRestartCompleted
    type: Running
```

For further LogPipeline examples, see the [samples](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/config/samples) directory.

## Custom resource parameters

### spec attribute

For details, see the [LogPipeline specification file](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/apis/telemetry/v1alpha1/logpipeline_types.go).

| Parameter | Type | Description |
|---|---|---|
| input | object | Definition where to collect logs, including selector mechanisms. |
| input.application | object | Input type for application logs collection. |
| input.application.namespaces | object | Provides selectors for Namespaces. Selectors are mutually exclusive. |
| input.application.namespaces.include | []string | List of Namespaces from which logs are collected. |
| input.application.namespaces.exclude | []string | List of Namespaces to exclude during log collection from all Namespaces. |
| input.application.namespaces.system | boolean | Set to `true` if collecting from all Namespaces must also include system Namespaces. |
| input.application.containers | []string | Provides selectors for containers. Selectors are mutually exclusive. |
| input.application.containers.include | []string | List of containers to collect from. |
| input.application.containers.exclude | []string | List of containers to exclude. |
| input.application.keepAnnotations | boolean | Indicates whether to keep all Kubernetes annotations. Default is `false`. |
| input.application.dropLabels | boolean | Indicates whether to drop all Kubernetes labels. Default is `false`. |
| filters | []object | List of [Fluent Bit filters](https://docs.fluentbit.io/manual/pipeline/filters) to apply to the logs processed by the pipeline. Filters are executed in sequence, as defined. They are executed before logs are buffered, and with that, are not executed on retries.|
| filters[].custom | string | Filter definition in the Fluent Bit syntax. **Note:** If you use a `custom` output, you put the LogPipeline in [unsupported mode](./../../01-overview/main-areas/telemetry/telemetry-02-logs.md#unsupported-mode).|
| output | object | [Fluent Bit output](https://docs.fluentbit.io/manual/pipeline/outputs) where you want to push the logs. Only one output can be specified. |
| output.grafana-loki | object | [Fluent Bit grafana-loki output](https://grafana.com/docs/loki/v2.2.x/clients/fluentbit/). **Note:** This output is considered legacy and is only provided for backward compatibility with the [deprecated](https://kyma-project.io/blog/2022/11/2/loki-deprecation/) in-cluster Loki instance. It might not be compatible with the latest Loki versions. For integration with a custom Loki installation use the `custom` output with the name `loki` instead, see also [this tutorial](https://github.com/kyma-project/examples/tree/main/loki). |
| output.grafana-loki.url | object | Grafana Loki URL. |
| output.grafana-loki.url.value | string | URL value. |
| output.grafana-loki.url.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| output.grafana-loki.labels | map[string]string | Labels to set for each log record. |
| output.grafana-loki.removeKeys | []string | Attributes to be removed from a log record. |
| output.http | object | Maps to a [Fluent Bit http output](https://docs.fluentbit.io/manual/pipeline/outputs/http). |
| output.http.compress | string | Payload compression mechanism. |
| output.http.dedot | boolean | If `true`, replaces dots with underscores ("dedotting") in the log field names `kubernetes.annotations` and `kubernetes.labels`. Default is `false`. |
| output.http.format | string | Data format to be used in the HTTP request body. Default is `json`. |
| output.http.host | object | IP address or hostname of the target HTTP server. |
| output.http.host.value | string | Host value, can contain references to Secret values. |
| output.http.host.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| output.http.password | object | Basic Auth password. |
| output.http.password.value | string | Password value, can contain references to Secret values. |
| output.http.password.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`.|
| output.http.port | string | TCP port of the target HTTP server. Default is `443`.  |
| output.http.tls | object | TLS Configuration of the HTTP target server.  |
| output.http.tls.disabled | boolean | Indicates if TLS is disabled or enabled. Default is `false`. |
| output.http.tls.skipCertificateValidation | boolean | If `true`, the validation of certificates is skipped. Default is `false`. |
| output.http.uri | string | URI for the target HTTP server. Fluent Bit Default is `/`. |
| output.http.user | object | Basic Auth username.|
| output.http.user.value | string | Username value, can contain references to Secret values. |
| output.http.user.valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| output.custom | string | Any other Fluent Bit output specified in the Fluent Bit configuration syntax. **Note:** If you use a `custom` output, you put the LogPipeline in [unsupported mode](./../../01-overview/main-areas/telemetry/telemetry-02-logs.md#unsupported-mode#unsupported-mode). |
| variables | []object | A list of mappings from Kubernetes Secret keys to environment variables. Mapped keys are mounted as environment variables, so that they are available as [Variables](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/variables) in the sections.|
| variables[].name | string | Name of the variable to map. |
| variables[].valueFrom.secretKeyRef | object | Reference to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`.|
| files | []object | A list of text snippets that are mounted as files to Fluent Bit, so that they are available for reference in filters and outputs. The mounted snippet is available under the `/files` folder.|
| files[].name | string | The file name under which the snippet is mounted. The resulting path will be `/files/<name>`. |
| files[].content | string | The actual text snippet to mount as file.|


### status attribute

For details, see the [LogPipeline specification file](https://github.com/kyma-project/kyma/blob/main/components/telemetry-operator/apis/telemetry/v1alpha1/logpipeline_types.go).

| Parameter | Type | Description |
|---|---|---|
| conditions | []object | An array of conditions describing the status of the pipeline.
| conditions[].lastTransitionTime | []object | An array of conditions describing the status of the pipeline.
| conditions[].reason | []object | An array of conditions describing the status of the pipeline.
| conditions[].type | enum | The possible transition types are:<br>- `Running`: The instance is ready and usable.<br>- `Pending`: The pipeline is being activated. |
| unsupportedMode | bool | Is active when the LogPipeline uses a `custom` output or filter; see [unsupported mode](./../../01-overview/main-areas/telemetry/telemetry-02-logs.md#unsupported-mode#unsupported-mode).
