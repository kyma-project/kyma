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
### LogPipeline.telemetry.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **files**  | \[\]object | Provides file content to be consumed by a LogPipeline configuration |
| **files.&#x200b;content**  | string |  |
| **files.&#x200b;name**  | string |  |
| **filters**  | \[\]object | Describes a filtering option on the logs of the pipeline. |
| **filters.&#x200b;custom**  | string | Custom filter definition in the Fluent Bit syntax. Note: If you use a `custom` filter, you put the LogPipeline in unsupported mode. |
| **input**  | object | Defines where to collect logs, including selector mechanisms. |
| **input.&#x200b;application**  | object | Configures in more detail from which containers application logs are enabled as input. |
| **input.&#x200b;application.&#x200b;containers**  | object | Describes whether application logs from specific containers are selected. The options are mutually exclusive. |
| **input.&#x200b;application.&#x200b;containers.&#x200b;exclude**  | \[\]string | Specifies to exclude only the container logs with the specified container names. |
| **input.&#x200b;application.&#x200b;containers.&#x200b;include**  | \[\]string | Specifies to include only the container logs with the specified container names. |
| **input.&#x200b;application.&#x200b;dropLabels**  | boolean | Defines whether to drop all Kubernetes labels. The default is `false`. |
| **input.&#x200b;application.&#x200b;keepAnnotations**  | boolean | Defines whether to keep all Kubernetes annotations. The default is `false`. |
| **input.&#x200b;application.&#x200b;namespaces**  | object | Describes whether application logs from specific Namespaces are selected. The options are mutually exclusive. System Namespaces are excluded by default from the collection. |
| **input.&#x200b;application.&#x200b;namespaces.&#x200b;exclude**  | \[\]string | Exclude the container logs of the specified Namespace names. |
| **input.&#x200b;application.&#x200b;namespaces.&#x200b;include**  | \[\]string | Include only the container logs of the specified Namespace names. |
| **input.&#x200b;application.&#x200b;namespaces.&#x200b;system**  | boolean | Set to `true` if collecting from all Namespaces must also include the system Namespaces like kube-system, istio-system, and kyma-system. |
| **output**  | object | [Fluent Bit output](https://docs.fluentbit.io/manual/pipeline/outputs) where you want to push the logs. Only one output can be specified. |
| **output.&#x200b;custom**  | string | Defines a custom output in the Fluent Bit syntax. Note: If you use a `custom` output, you put the LogPipeline in unsupported mode. |
| **output.&#x200b;grafana-loki**  | object | Configures an output to the Kyma-internal Loki instance. [Fluent Bit grafana-loki output](https://grafana.com/docs/loki/v2.2.x/clients/fluentbit/). **Note:** This output is considered legacy and is only provided for backward compatibility with the [deprecated](https://kyma-project.io/blog/2022/11/2/loki-deprecation/) in-cluster Loki instance. It might not be compatible with the latest Loki versions. For integration with a custom Loki installation use the `custom` output with the name `loki` instead, see also [Installing a custom Loki stack in Kyma](https://github.com/kyma-project/examples/tree/main/loki). |
| **output.&#x200b;grafana-loki.&#x200b;labels**  | map\[string\]string | Labels to set for each log record. |
| **output.&#x200b;grafana-loki.&#x200b;removeKeys**  | \[\]string | Attributes to be removed from a log record. |
| **output.&#x200b;grafana-loki.&#x200b;url**  | object | Grafana Loki URL. |
| **output.&#x200b;grafana-loki.&#x200b;url.&#x200b;value**  | string | Value that can contain references to Secret values. |
| **output.&#x200b;grafana-loki.&#x200b;url.&#x200b;valueFrom**  | object |  |
| **output.&#x200b;grafana-loki.&#x200b;url.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;grafana-loki.&#x200b;url.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key**  | string |  |
| **output.&#x200b;grafana-loki.&#x200b;url.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name**  | string |  |
| **output.&#x200b;grafana-loki.&#x200b;url.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace**  | string |  |
| **output.&#x200b;http**  | object | Configures an HTTP-based output compatible with the Fluent Bit HTTP output plugin. |
| **output.&#x200b;http.&#x200b;compress**  | string | Defines the compression algorithm to use. |
| **output.&#x200b;http.&#x200b;dedot**  | boolean | Enables de-dotting of Kubernetes labels and annotations for compatibility with ElasticSearch based backends. Dots (.) will be replaced by underscores (_). Default is `false`. |
| **output.&#x200b;http.&#x200b;format**  | string | Data format to be used in the HTTP request body. Default is `json`. |
| **output.&#x200b;http.&#x200b;host**  | object | Defines the host of the HTTP receiver. |
| **output.&#x200b;http.&#x200b;host.&#x200b;value**  | string | Value that can contain references to Secret values. |
| **output.&#x200b;http.&#x200b;host.&#x200b;valueFrom**  | object |  |
| **output.&#x200b;http.&#x200b;host.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;http.&#x200b;host.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key**  | string |  |
| **output.&#x200b;http.&#x200b;host.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name**  | string |  |
| **output.&#x200b;http.&#x200b;host.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace**  | string |  |
| **output.&#x200b;http.&#x200b;password**  | object | Defines the basic auth password. |
| **output.&#x200b;http.&#x200b;password.&#x200b;value**  | string | Value that can contain references to Secret values. |
| **output.&#x200b;http.&#x200b;password.&#x200b;valueFrom**  | object |  |
| **output.&#x200b;http.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;http.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key**  | string |  |
| **output.&#x200b;http.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name**  | string |  |
| **output.&#x200b;http.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace**  | string |  |
| **output.&#x200b;http.&#x200b;port**  | string | Defines the port of the HTTP receiver. Default is 443. |
| **output.&#x200b;http.&#x200b;tls**  | object | Configures TLS for the HTTP target server. |
| **output.&#x200b;http.&#x200b;tls.&#x200b;disabled**  | boolean | Indicates if TLS is disabled or enabled. Default is `false`. |
| **output.&#x200b;http.&#x200b;tls.&#x200b;skipCertificateValidation**  | boolean | If `true`, the validation of certificates is skipped. Default is `false`. |
| **output.&#x200b;http.&#x200b;uri**  | string | Defines the URI of the HTTP receiver. Default is "/". |
| **output.&#x200b;http.&#x200b;user**  | object | Defines the basic auth user. |
| **output.&#x200b;http.&#x200b;user.&#x200b;value**  | string | Value that can contain references to Secret values. |
| **output.&#x200b;http.&#x200b;user.&#x200b;valueFrom**  | object |  |
| **output.&#x200b;http.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;http.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key**  | string |  |
| **output.&#x200b;http.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name**  | string |  |
| **output.&#x200b;http.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace**  | string |  |
| **variables**  | \[\]object | A list of mappings from Kubernetes Secret keys to environment variables. Mapped keys are mounted as environment variables, so that they are available as [Variables](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/variables) in the sections. |
| **variables.&#x200b;name**  | string | Name of the variable to map. |
| **variables.&#x200b;valueFrom**  | object |  |
| **variables.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **variables.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key**  | string |  |
| **variables.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name**  | string |  |
| **variables.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace**  | string |  |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | An array of conditions describing the status of the pipeline. |
| **conditions.&#x200b;lastTransitionTime**  | string | An array of conditions describing the status of the pipeline. |
| **conditions.&#x200b;reason**  | string | An array of conditions describing the status of the pipeline. |
| **conditions.&#x200b;type**  | string | The possible transition types are:<br>- `Running`: The instance is ready and usable.<br>- `Pending`: The pipeline is being activated. |
| **unsupportedMode**  | boolean | Is active when the LogPipeline uses a `custom` output or filter; see [unsupported mode](../../01-overview/telemetry/telemetry-02-logs.md#unsupported-mode). |

<!-- TABLE-END -->