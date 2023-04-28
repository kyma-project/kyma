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
### LogPipeline.v1alpha1.telemetry.kyma-project.io

**Spec:**

<!-- LogPipeline v1alpha1 telemetry.kyma-project.io -->
| Parameter         | Type | Description                                   |
| ------------------| ---- | --------------------------------------------- |
| **** | object | Defines the desired state of LogPipeline |
| **files** | Provides file content to be consumed by a LogPipeline configuration | files |
| **files.content** | string |  |
| **files.name** | string |  |
| **filters** | Describes a filtering option on the logs of the pipeline. | filters |
| **filters.custom** | string | Custom filter definition in the Fluent Bit syntax. Note: If you use a `custom` filter, you put the LogPipeline in unsupported mode. |
| **input** | object | Defines where to collect logs, including selector mechanisms. |
| **input.application** | object | Configures in more detail from which containers application logs are enabled as input. |
| **input.application.containers** | object | Describes whether application logs from specific containers are selected. The options are mutually exclusive. |
| **input.application.containers.exclude** | Specifies to exclude only the container logs with the specified container names. | exclude |
| **input.application.containers.include** | Specifies to include only the container logs with the specified container names. | include |
| **input.application.dropLabels** | boolean | Defines whether to drop all Kubernetes labels. The default is `false`. |
| **input.application.keepAnnotations** | boolean | Defines whether to keep all Kubernetes annotations. The default is `false`. |
| **input.application.namespaces** | object | Describes whether application logs from specific Namespaces are selected. The options are mutually exclusive. System Namespaces are excluded by default from the collection. |
| **input.application.namespaces.exclude** | Exclude the container logs of the specified Namespace names. | exclude |
| **input.application.namespaces.include** | Include only the container logs of the specified Namespace names. | include |
| **input.application.namespaces.system** | boolean | Set to `true` if collecting from all Namespaces must also include the system Namespaces like kube-system, istio-system, and kyma-system. |
| **output** | object | [Fluent Bit output](https://docs.fluentbit.io/manual/pipeline/outputs) where you want to push the logs. Only one output can be specified. |
| **output.custom** | string | Defines a custom output in the Fluent Bit syntax. Note: If you use a `custom` output, you put the LogPipeline in unsupported mode. |
| **output.grafana-loki** | object | Configures an output to the Kyma-internal Loki instance. [Fluent Bit grafana-loki output](https://grafana.com/docs/loki/v2.2.x/clients/fluentbit/). **Note:** This output is considered legacy and is only provided for backward compatibility with the [deprecated](https://kyma-project.io/blog/2022/11/2/loki-deprecation/) in-cluster Loki instance. It might not be compatible with the latest Loki versions. For integration with a custom Loki installation use the `custom` output with the name `loki` instead, see also [Installing a custom Loki stack in Kyma](https://github.com/kyma-project/examples/tree/main/loki). |
| **output.grafana-loki.labels** | object | Labels to set for each log record. |
| **output.grafana-loki.removeKeys** | Attributes to be removed from a log record. | removeKeys |
| **output.grafana-loki.url** | object | Grafana Loki URL. |
| **output.grafana-loki.url.value** | string | Value that can contain references to Secret values. |
| **output.grafana-loki.url.valueFrom** | object |  |
| **output.grafana-loki.url.valueFrom.secretKeyRef** | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.grafana-loki.url.valueFrom.secretKeyRef.key** | string |  |
| **output.grafana-loki.url.valueFrom.secretKeyRef.name** | string |  |
| **output.grafana-loki.url.valueFrom.secretKeyRef.namespace** | string |  |
| **output.http** | object | Configures an HTTP-based output compatible with the Fluent Bit HTTP output plugin. |
| **output.http.compress** | string | Defines the compression algorithm to use. |
| **output.http.dedot** | boolean | Enables de-dotting of Kubernetes labels and annotations for compatibility with ElasticSearch based backends. Dots (.) will be replaced by underscores (_). Default is `false`. |
| **output.http.format** | string | Data format to be used in the HTTP request body. Default is `json`. |
| **output.http.host** | object | Defines the host of the HTTP receiver. |
| **output.http.host.value** | string | Value that can contain references to Secret values. |
| **output.http.host.valueFrom** | object |  |
| **output.http.host.valueFrom.secretKeyRef** | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.http.host.valueFrom.secretKeyRef.key** | string |  |
| **output.http.host.valueFrom.secretKeyRef.name** | string |  |
| **output.http.host.valueFrom.secretKeyRef.namespace** | string |  |
| **output.http.password** | object | Defines the basic auth password. |
| **output.http.password.value** | string | Value that can contain references to Secret values. |
| **output.http.password.valueFrom** | object |  |
| **output.http.password.valueFrom.secretKeyRef** | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.http.password.valueFrom.secretKeyRef.key** | string |  |
| **output.http.password.valueFrom.secretKeyRef.name** | string |  |
| **output.http.password.valueFrom.secretKeyRef.namespace** | string |  |
| **output.http.port** | string | Defines the port of the HTTP receiver. Default is 443. |
| **output.http.tls** | object | Configures TLS for the HTTP target server. |
| **output.http.tls.disabled** | boolean | Indicates if TLS is disabled or enabled. Default is `false`. |
| **output.http.tls.skipCertificateValidation** | boolean | If `true`, the validation of certificates is skipped. Default is `false`. |
| **output.http.uri** | string | Defines the URI of the HTTP receiver. Default is "/". |
| **output.http.user** | object | Defines the basic auth user. |
| **output.http.user.value** | string | Value that can contain references to Secret values. |
| **output.http.user.valueFrom** | object |  |
| **output.http.user.valueFrom.secretKeyRef** | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.http.user.valueFrom.secretKeyRef.key** | string |  |
| **output.http.user.valueFrom.secretKeyRef.name** | string |  |
| **output.http.user.valueFrom.secretKeyRef.namespace** | string |  |
| **variables** | A list of mappings from Kubernetes Secret keys to environment variables. Mapped keys are mounted as environment variables, so that they are available as [Variables](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/variables) in the sections. | variables |
| **variables.name** | string | Name of the variable to map. |
| **variables.valueFrom** | object |  |
| **variables.valueFrom.secretKeyRef** | object | Refers to a key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **variables.valueFrom.secretKeyRef.key** | string |  |
| **variables.valueFrom.secretKeyRef.name** | string |  |
| **variables.valueFrom.secretKeyRef.namespace** | string |  |

**Status:**

<!-- LogPipeline v1alpha1 telemetry.kyma-project.io -->
| Parameter         | Type | Description                                   |
| ------------------| ---- | --------------------------------------------- |
| **** | object | Shows the observed state of the LogPipeline |
| **conditions** | An array of conditions describing the status of the pipeline. | conditions |
| **conditions.lastTransitionTime** | string | An array of conditions describing the status of the pipeline. |
| **conditions.reason** | string | An array of conditions describing the status of the pipeline. |
| **conditions.type** | string | The possible transition types are:<br>- `Running`: The instance is ready and usable.<br>- `Pending`: The pipeline is being activated. |
| **unsupportedMode** | boolean | Is active when the LogPipeline uses a `custom` output or filter; see [unsupported mode](./../../01-overview/main-areas/telemetry/telemetry-02-logs.md#unsupported-mode#unsupported-mode). |


<!-- TABLE-END -->