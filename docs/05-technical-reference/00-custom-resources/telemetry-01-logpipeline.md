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

For further LogPipeline examples, see the [samples](https://github.com/kyma-project/telemetry-manager/tree/main/config/samples) directory.

## Custom resource parameters

For details, see the [LogPipeline specification file](https://github.com/kyma-project/telemetry-manager/blob/main/apis/telemetry/v1alpha1/logpipeline_types.go).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
<!-- LogPipeline v1alpha1 telemetry.kyma-project.io -->
| Parameter         | Description                                   |
| ---------------------------------------- | ---------|
| **spec.files** | Provides file content to be consumed by a LogPipeline configuration |
| **spec.files.content** |  |
| **spec.files.name** |  |
| **spec.filters** | Describes a filtering option on the logs of the pipeline. |
| **spec.filters.custom** | Custom filter definition in the Fluent Bit syntax. Note: If you use a `custom` filter, you put the LogPipeline in unsupported mode. |
| **spec.input** | Defines where to collect logs, including selector mechanisms. |
| **spec.input.application** | Configures in more detail from which containers application logs are enabled as input. |
| **spec.input.application.containers** | Describes whether application logs from specific containers are selected. The options are mutually exclusive. |
| **spec.input.application.containers.exclude** | Specifies to exclude only the container logs with the specified container names. |
| **spec.input.application.containers.include** | Specifies to include only the container logs with the specified container names. qwe. |
| **spec.input.application.dropLabels** | Defines whether to drop all Kubernetes labels. The default is `false`. |
| **spec.input.application.keepAnnotations** | Defines whether to keep all Kubernetes annotations. The default is `false`. |
| **spec.input.application.namespaces** | Describes whether application logs from specific Namespaces are selected. The options are mutually exclusive. System Namespaces are excluded by default from the collection. |
| **spec.input.application.namespaces.exclude** | Exclude the container logs of the specified Namespace names. |
| **spec.input.application.namespaces.include** | Include only the container logs of the specified Namespace names. |
| **spec.input.application.namespaces.system** | Set to `true` if collecting from all Namespaces must also include the system Namespaces like kube-system, istio-system, and kyma-system. |
| **spec.output** | [Fluent Bit output](https://docs.fluentbit.io/manual/pipeline/outputs) where you want to push the logs. Only one output can be specified. |
| **spec.output.custom** | Defines a custom output in the Fluent Bit syntax. Note: If you use a `custom` output, you put the LogPipeline in unsupported mode. |
| **spec.output.grafana-loki** | Configures an output to the Kyma-internal Loki instance. [Fluent Bit grafana-loki output](https://grafana.com/docs/loki/v2.2.x/clients/fluentbit/). **Note:** This output is considered legacy and is only provided for backward compatibility with the [deprecated](https://kyma-project.io/blog/2022/11/2/loki-deprecation/) in-cluster Loki instance. It might not be compatible with the latest Loki versions. For integration with a custom Loki installation use the `custom` output with the name `loki` instead, see also [Installing a custom Loki stack in Kyma](https://github.com/kyma-project/examples/tree/main/loki). |
| **spec.output.grafana-loki.labels** | Labels to set for each log record. |
| **spec.output.grafana-loki.removeKeys** | Attributes to be removed from a log record. |
| **spec.output.grafana-loki.url** | Grafana Loki URL. |
| **spec.output.grafana-loki.url.value** | Value that can contain references to Secret values. |
| **spec.output.grafana-loki.url.valueFrom** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef** | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef.key** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef.name** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.http** | Configures an HTTP-based output compatible with the Fluent Bit HTTP output plugin. |
| **spec.output.http.compress** | Defines the compression algorithm to use. |
| **spec.output.http.dedot** | Enables de-dotting of Kubernetes labels and annotations for compatibility with ElasticSearch based backends. Dots (.) will be replaced by underscores (_). Default is `false`. |
| **spec.output.http.format** | Data format to be used in the HTTP request body. Default is `json`. |
| **spec.output.http.host** | Defines the host of the HTTP receiver. |
| **spec.output.http.host.value** | Value that can contain references to Secret values. |
| **spec.output.http.host.valueFrom** |  |
| **spec.output.http.host.valueFrom.secretKeyRef** | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **spec.output.http.host.valueFrom.secretKeyRef.key** |  |
| **spec.output.http.host.valueFrom.secretKeyRef.name** |  |
| **spec.output.http.host.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.http.password** | Defines the basic auth password. |
| **spec.output.http.password.value** | Value that can contain references to Secret values. |
| **spec.output.http.password.valueFrom** |  |
| **spec.output.http.password.valueFrom.secretKeyRef** | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **spec.output.http.password.valueFrom.secretKeyRef.key** |  |
| **spec.output.http.password.valueFrom.secretKeyRef.name** |  |
| **spec.output.http.password.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.http.port** | Defines the port of the HTTP receiver. Default is 443. |
| **spec.output.http.tls** | Configures TLS for the HTTP target server. |
| **spec.output.http.tls.disabled** | Indicates if TLS is disabled or enabled. Default is `false`. |
| **spec.output.http.tls.skipCertificateValidation** | If `true`, the validation of certificates is skipped. Default is `false`. |
| **spec.output.http.uri** | Defines the URI of the HTTP receiver. Default is "/". |
| **spec.output.http.user** | Defines the basic auth user. |
| **spec.output.http.user.value** | Value that can contain references to Secret values. |
| **spec.output.http.user.valueFrom** |  |
| **spec.output.http.user.valueFrom.secretKeyRef** | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **spec.output.http.user.valueFrom.secretKeyRef.key** |  |
| **spec.output.http.user.valueFrom.secretKeyRef.name** |  |
| **spec.output.http.user.valueFrom.secretKeyRef.namespace** |  |
| **spec.variables** | A list of mappings from Kubernetes Secret keys to environment variables. Mapped keys are mounted as environment variables, so that they are available as [Variables](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/variables) in the sections. |
| **spec.variables.name** | Name of the variable to map. |
| **spec.variables.valueFrom** |  |
| **spec.variables.valueFrom.secretKeyRef** | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **spec.variables.valueFrom.secretKeyRef.key** |  |
| **spec.variables.valueFrom.secretKeyRef.name** |  |
| **spec.variables.valueFrom.secretKeyRef.namespace** |  |
| **status.conditions** | An array of conditions describing the status of the pipeline. |
| **status.conditions.lastTransitionTime** | An array of conditions describing the status of the pipeline. |
| **status.conditions.reason** | An array of conditions describing the status of the pipeline. |
| **status.conditions.type** | The possible transition types are:<br>- `Running`: The instance is ready and usable.<br>- `Pending`: The pipeline is being activated. |
| **status.unsupportedMode** | Is active when the LogPipeline uses a `custom` output or filter; see [unsupported mode](./../../01-overview/main-areas/telemetry/telemetry-02-logs.md#unsupported-mode#unsupported-mode). |<!-- TABLE-END -->