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

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- SKIP-WITH-ANCESTORS spec.template -->

<!-- TABLE-START -->
<!-- LogPipeline -->
| Parameter         | Description                                   |
| ---------------------------------------- | ---------|
| **spec.files** | Provides file content to be consumed by a LogPipeline configuration |
| **spec.files.content** |  |
| **spec.files.name** |  |
| **spec.filters** | Describes a filtering option on the logs of the pipeline. |
| **spec.filters.custom** | Filter definition in the Fluent Bit syntax. Note: If you use a `custom` filter, you put the LogPipeline in unsupported mode. |
| **spec.input** | Describes a log input for a LogPipeline. |
| **spec.input.application** | Configures in more detail from which containers application logs are enabled as input. |
| **spec.input.application.containers** | Describes whether application logs from specific containers are selected. The options are mutually exclusive. |
| **spec.input.application.containers.exclude** | Specifies to exclude only the container logs with the specified container names. |
| **spec.input.application.containers.include** | Specifies to include only the container logs with the specified container names. |
| **spec.input.application.dropLabels** | Defines whether to drop all Kubernetes labels. The default is false. |
| **spec.input.application.keepAnnotations** | Defines whether to keep all Kubernetes annotations. The default is false. |
| **spec.input.application.namespaces** | Describes whether application logs from specific Namespaces are selected. The options are mutually exclusive. System Namespaces are excluded by default from the collection. |
| **spec.input.application.namespaces.exclude** | The container logs of the specified Namespace names. |
| **spec.input.application.namespaces.include** | Only the container logs of the specified Namespace names. |
| **spec.input.application.namespaces.system** | Describes to include the container logs of the system Namespaces like kube-system, istio-system, and kyma-system. |
| **spec.output** | Describes a Fluent Bit output configuration section. |
| **spec.output.custom** | Defines a custom output in the Fluent Bit syntax. Note: If you use a `custom` output, you put the LogPipeline in unsupported mode. |
| **spec.output.grafana-loki** | Configures an output to the Kyma-internal Loki instance. Note: This output is considered legacy and is only provided for backwards compatibility with the in-cluster Loki instance. It might not be compatible with latest Loki versions. For integration with a Loki-based system, use the `custom` output with name `loki` instead. |
| **spec.output.grafana-loki.labels** |  |
| **spec.output.grafana-loki.removeKeys** |  |
| **spec.output.grafana-loki.url** |  |
| **spec.output.grafana-loki.url.value** |  |
| **spec.output.grafana-loki.url.valueFrom** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef.key** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef.name** |  |
| **spec.output.grafana-loki.url.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.http** | Configures an HTTP-based output compatible with the Fluent Bit HTTP output plugin. |
| **spec.output.http.compress** | Defines the compression algorithm to use. |
| **spec.output.http.dedot** | Enables de-dotting of Kubernetes labels and annotations for compatibility with ElasticSearch based backends. Dots (.) will be replaced by underscores (_). |
| **spec.output.http.format** | Defines the log encoding to be used. Default is json. |
| **spec.output.http.host** | Defines the host of the HTTP receiver. |
| **spec.output.http.host.value** |  |
| **spec.output.http.host.valueFrom** |  |
| **spec.output.http.host.valueFrom.secretKeyRef** |  |
| **spec.output.http.host.valueFrom.secretKeyRef.key** |  |
| **spec.output.http.host.valueFrom.secretKeyRef.name** |  |
| **spec.output.http.host.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.http.password** | Defines the basic auth password. |
| **spec.output.http.password.value** |  |
| **spec.output.http.password.valueFrom** |  |
| **spec.output.http.password.valueFrom.secretKeyRef** |  |
| **spec.output.http.password.valueFrom.secretKeyRef.key** |  |
| **spec.output.http.password.valueFrom.secretKeyRef.name** |  |
| **spec.output.http.password.valueFrom.secretKeyRef.namespace** |  |
| **spec.output.http.port** | Defines the port of the HTTP receiver. Default is 443. |
| **spec.output.http.tls** | Defines TLS settings for the HTTP connection. |
| **spec.output.http.tls.disabled** | Disable TLS. |
| **spec.output.http.tls.skipCertificateValidation** | Disable TLS certificate validation. |
| **spec.output.http.uri** | Defines the URI of the HTTP receiver. Default is "/". |
| **spec.output.http.user** | Defines the basic auth user. |
| **spec.output.http.user.value** |  |
| **spec.output.http.user.valueFrom** |  |
| **spec.output.http.user.valueFrom.secretKeyRef** |  |
| **spec.output.http.user.valueFrom.secretKeyRef.key** |  |
| **spec.output.http.user.valueFrom.secretKeyRef.name** |  |
| **spec.output.http.user.valueFrom.secretKeyRef.namespace** |  |
| **spec.variables** | References a Kubernetes secret that should be provided as environment variable to Fluent Bit |
| **spec.variables.name** |  |
| **spec.variables.valueFrom** |  |
| **spec.variables.valueFrom.secretKeyRef** |  |
| **spec.variables.valueFrom.secretKeyRef.key** |  |
| **spec.variables.valueFrom.secretKeyRef.name** |  |
| **spec.variables.valueFrom.secretKeyRef.namespace** |  |
| **status.conditions** | LogPipelineCondition contains details for the current condition of this LogPipeline |
| **status.conditions.lastTransitionTime** |  |
| **status.conditions.reason** |  |
| **status.conditions.type** |  |
| **status.unsupportedMode** |  |<!-- TABLE-END -->