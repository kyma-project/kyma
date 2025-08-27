# MetricPipeline

The `metricpipeline.telemetry.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to filter and ship metric data in Kyma. To get the current CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd metricpipeline.telemetry.kyma-project.io -o yaml
```

## Sample Custom Resource

The following MetricPipeline object defines a pipeline that integrates into an OTLP backend:

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: MetricPipeline
metadata:
  name: otlp
  generation: 1
spec:
  input:
    application:
      prometheus:
        enabled: false
      istio:
        enabled: false
      runtime:
        enabled: false
  output:
    otlp:
      endpoint:
        value: https://myBackend:4317
status:
  conditions:
  - lastTransitionTime: "2024-01-09T07:02:16Z"
    message: "Metric agent DaemonSet is ready"
    observedGeneration: 1
    reason: AgentReady
    status: "True"
    type: AgentHealthy
  - lastTransitionTime: "2024-01-08T10:40:18Z"
    message: "Metric gateway Deployment is ready"
    observedGeneration: 1
    reason: GatewayReady
    status: "True"
    type: GatewayHealthy
  - lastTransitionTime: "2023-12-28T11:27:04Z"
    message: ""
    observedGeneration: 1
    reason: ConfigurationGenerated
    status: "True"
    type: ConfigurationGenerated
```

For further examples, see the [samples](https://github.com/kyma-project/telemetry-manager/tree/main/config/samples) directory.

## Custom Resource Parameters

For details, see the [MetricPipeline specification file](https://github.com/kyma-project/telemetry-manager/blob/main/apis/telemetry/v1alpha1/metricpipeline_types.go).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
### MetricPipeline.telemetry.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **input**  | object | Input configures additional inputs for metric collection. |
| **input.&#x200b;istio**  | object | Istio input configures collection of Istio metrics from applications running in the Istio service mesh. |
| **input.&#x200b;istio.&#x200b;diagnosticMetrics**  | object | DiagnosticMetrics configures collection of additional diagnostic metrics. The default is `false`. |
| **input.&#x200b;istio.&#x200b;diagnosticMetrics.&#x200b;enabled**  | boolean | If enabled, diagnostic metrics are collected. The default is `false`. |
| **input.&#x200b;istio.&#x200b;enabled**  | boolean | Enabled specifies that istio-proxy metrics are scraped from Pods that have the istio-proxy sidecar injected. The default is `false`. |
| **input.&#x200b;istio.&#x200b;envoyMetrics**  | object | EnvoyMetrics enables the collection of additional Envoy metrics with prefix `envoy_`. The default is `false`. |
| **input.&#x200b;istio.&#x200b;envoyMetrics.&#x200b;enabled**  | boolean | Enabled specifies that Envoy metrics with prefix `envoy_` are scraped additionally. The default is `false`. |
| **input.&#x200b;istio.&#x200b;namespaces**  | object | Namespaces configures the namespaces for which the collection should be activated. By default, all namespaces including system namespaces are enabled. |
| **input.&#x200b;istio.&#x200b;namespaces.&#x200b;exclude**  | \[\]string | Exclude signals from the specified Namespace names only. |
| **input.&#x200b;istio.&#x200b;namespaces.&#x200b;include**  | \[\]string | Include signals from the specified Namespace names only. |
| **input.&#x200b;otlp**  | object | OTLP input configures the push endpoint to receive metrics from an OTLP source. |
| **input.&#x200b;otlp.&#x200b;disabled**  | boolean | If set to `true`, no push-based OTLP signals are collected. The default is `false`. |
| **input.&#x200b;otlp.&#x200b;namespaces**  | object | Namespaces describes whether push-based OTLP signals from specific namespaces are selected. System namespaces are enabled by default. |
| **input.&#x200b;otlp.&#x200b;namespaces.&#x200b;exclude**  | \[\]string | Exclude signals from the specified Namespace names only. |
| **input.&#x200b;otlp.&#x200b;namespaces.&#x200b;include**  | \[\]string | Include signals from the specified Namespace names only. |
| **input.&#x200b;prometheus**  | object | Prometheus input configures collection of application metrics in the pull-based Prometheus protocol using endpoint discovery based on annotations. |
| **input.&#x200b;prometheus.&#x200b;diagnosticMetrics**  | object | DiagnosticMetrics configures collection of additional diagnostic metrics. The default is `false`. |
| **input.&#x200b;prometheus.&#x200b;diagnosticMetrics.&#x200b;enabled**  | boolean | If enabled, diagnostic metrics are collected. The default is `false`. |
| **input.&#x200b;prometheus.&#x200b;enabled**  | boolean | Enabled specifies whether Service endpoints and Pods marked with `prometheus.io/scrape=true` annotation are scraped. The default is `false`. |
| **input.&#x200b;prometheus.&#x200b;namespaces**  | object | Namespaces specifies from which namespaces metrics are collected. By default, all namespaces except the system namespaces are enabled. To enable all namespaces including system namespaces, use an empty struct notation. |
| **input.&#x200b;prometheus.&#x200b;namespaces.&#x200b;exclude**  | \[\]string | Exclude signals from the specified Namespace names only. |
| **input.&#x200b;prometheus.&#x200b;namespaces.&#x200b;include**  | \[\]string | Include signals from the specified Namespace names only. |
| **input.&#x200b;runtime**  | object | Runtime input configures collection of Kubernetes runtime metrics. |
| **input.&#x200b;runtime.&#x200b;enabled**  | boolean | Enabled specifies whether runtime metrics are collected. The default is `false`. |
| **input.&#x200b;runtime.&#x200b;namespaces**  | object | Namespaces specifies from which namespaces metrics are collected. By default, all namespaces except the system namespaces are enabled. To enable all namespaces including system namespaces, use an empty struct notation. |
| **input.&#x200b;runtime.&#x200b;namespaces.&#x200b;exclude**  | \[\]string | Exclude signals from the specified Namespace names only. |
| **input.&#x200b;runtime.&#x200b;namespaces.&#x200b;include**  | \[\]string | Include signals from the specified Namespace names only. |
| **input.&#x200b;runtime.&#x200b;resources**  | object | Resources configures the Kubernetes resource types for which metrics are collected. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;container**  | object | Container configures container runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;container.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;daemonset**  | object | DaemonSet configures DaemonSet runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;daemonset.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;deployment**  | object | Deployment configures Deployment runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;deployment.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;job**  | object | Job configures Job runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;job.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;node**  | object | Node configures Node runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;node.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;pod**  | object | Pod configures Pod runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;pod.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;statefulset**  | object | StatefulSet configures StatefulSet runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;statefulset.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;volume**  | object | Volume configures Volume runtime metrics collection. |
| **input.&#x200b;runtime.&#x200b;resources.&#x200b;volume.&#x200b;enabled**  | boolean | Enabled specifies that the runtime metrics for the resource type are collected. The default is `true`. |
| **output**  | object | Output configures the backend to which metrics are sent. You must specify exactly one output per pipeline. |
| **output.&#x200b;otlp** (required) | object | OTLP output defines an output using the OpenTelemetry protocol. |
| **output.&#x200b;otlp.&#x200b;authentication**  | object | Authentication defines authentication options for the OTLP output |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic**  | object | Basic activates `Basic` authentication for the destination providing relevant Secrets. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password** (required) | object | Password contains the basic auth password or a Secret reference. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;password.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user** (required) | object | User contains the basic auth username or a Secret reference. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;authentication.&#x200b;basic.&#x200b;user.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |
| **output.&#x200b;otlp.&#x200b;endpoint** (required) | object | Endpoint defines the host and port (`<host>:<port>`) of an OTLP endpoint. |
| **output.&#x200b;otlp.&#x200b;endpoint.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;endpoint.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;endpoint.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;endpoint.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;endpoint.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;endpoint.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |
| **output.&#x200b;otlp.&#x200b;headers**  | \[\]object | Headers defines custom headers to be added to outgoing HTTP or gRPC requests. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;name** (required) | string | Name defines the header name. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;prefix**  | string | Prefix defines an optional header value prefix. The prefix is separated from the value by a space character. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;headers.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |
| **output.&#x200b;otlp.&#x200b;path**  | string | Path defines OTLP export URL path (only for the HTTP protocol). This value overrides auto-appended paths `/v1/metrics` and `/v1/traces` |
| **output.&#x200b;otlp.&#x200b;protocol**  | string | Protocol defines the OTLP protocol (`http` or `grpc`). Default is `grpc`. |
| **output.&#x200b;otlp.&#x200b;tls**  | object | TLS defines TLS options for the OTLP output. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca**  | object | Defines an optional CA certificate for server certificate verification when using TLS. The certificate must be provided in PEM format. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;ca.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert**  | object | Defines a client certificate to use when using TLS. The certificate must be provided in PEM format. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;cert.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;insecure**  | boolean | Insecure defines whether to send requests using plaintext instead of TLS. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;insecureSkipVerify**  | boolean | InsecureSkipVerify defines whether to skip server certificate verification when using TLS. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key**  | object | Defines the client key to use when using TLS. The key must be provided in PEM format. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key.&#x200b;value**  | string | Value as plain text. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key.&#x200b;valueFrom**  | object | ValueFrom is the value as a reference to a resource. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | SecretKeyRef refers to the value of a specific key in a Secret. You must provide `name` and `namespace` of the Secret, as well as the name of the `key`. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | Key defines the name of the attribute of the Secret holding the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name** (required) | string | Name of the Secret containing the referenced value. |
| **output.&#x200b;otlp.&#x200b;tls.&#x200b;key.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;namespace** (required) | string | Namespace containing the Secret with the referenced value. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | An array of conditions describing the status of the pipeline. |
| **conditions.&#x200b;lastTransitionTime** (required) | string | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. |
| **conditions.&#x200b;message** (required) | string | message is a human readable message indicating details about the transition. This may be an empty string. |
| **conditions.&#x200b;observedGeneration**  | integer | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. |
| **conditions.&#x200b;reason** (required) | string | reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty. |
| **conditions.&#x200b;status** (required) | string | status of the condition, one of True, False, Unknown. |
| **conditions.&#x200b;type** (required) | string | type of condition in CamelCase or in foo.example.com/CamelCase. |

<!-- TABLE-END -->
### MetricPipeline Status

The status of the MetricPipeline is determined by the condition types `GatewayHealthy`, `AgentHealthy`, `ConfigurationGenerated`, and `TelemetryFlowHealthy`:

| Condition Type         | Condition Status | Condition Reason                | Condition Message                                                                                                                                                                                                                         |
|------------------------|------------------|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| GatewayHealthy         | True             | GatewayReady                    | Metric gateway Deployment is ready                                                                                                                                                                                                        |
| GatewayHealthy         | True             | RolloutInProgress               | Pods are being started/updated                                                                                                                                                                                                            |
| GatewayHealthy         | False            | GatewayNotReady                 | No Pods deployed                                                                                                                                                                                                                          |
| GatewayHealthy         | False            | GatewayNotReady                 | Failed to list ReplicaSets: `reason`                                                                                                                                                                                                      |
| GatewayHealthy         | False            | GatewayNotReady                 | Failed to fetch ReplicaSets: `reason`                                                                                                                                                                                                     |
| GatewayHealthy         | False            | GatewayNotReady                 | Pod is not scheduled: `reason`                                                                                                                                                                                                            |
| GatewayHealthy         | False            | GatewayNotReady                 | Pod is in the pending state because container: `container name` is not running due to: `reason`. Please check the container: `container name` logs.                                                                                       |
| GatewayHealthy         | False            | GatewayNotReady                 | Pod is in the failed state due to: `reason`                                                                                                                                                                                               |
| GatewayHealthy         | False            | GatewayNotReady                 | Deployment is not yet created                                                                                                                                                                                                             |
| GatewayHealthy         | False            | GatewayNotReady                 | Failed to get Deployment                                                                                                                                                                                                                  |
| GatewayHealthy         | False            | GatewayNotReady                 | Failed to get latest ReplicaSets                                                                                                                                                                                                          |
| AgentHealthy           | True             | AgentNotRequired                |                                                                                                                                                                                                                                           |
| AgentHealthy           | True             | AgentReady                      | Metric agent DaemonSet is ready                                                                                                                                                                                                           |
| AgentHealthy           | True             | RolloutInProgress               | Pods are being started/updated                                                                                                                                                                                                            |
| AgentHealthy           | False            | AgentNotReady                   | No Pods deployed                                                                                                                                                                                                                          |
| AgentHealthy           | False            | AgentNotReady                   | DaemonSet is not yet created                                                                                                                                                                                                              |
| AgentHealthy           | False            | AgentNotReady                   | Failed to get DaemonSet                                                                                                                                                                                                                   |
| AgentHealthy           | False            | AgentNotReady                   | Pod is in the pending state because container: `container name` is not running due to: `reason`                                                                                                                                           |
| AgentHealthy           | False            | AgentNotReady                   | Pod is in the failed state due to: `reason`                                                                                                                                                                                               |
| ConfigurationGenerated | True             | AgentGatewayConfigured          | MetricPipeline specification is successfully applied to the configuration of Metric gateway                                                                                                                                               |
| ConfigurationGenerated | True             | TLSCertificateAboutToExpire     | TLS (CA) certificate is about to expire, configured certificate is valid until YYYY-MM-DD                                                                                                                                                 |
| ConfigurationGenerated | False            | EndpointInvalid                 | OTLP output endpoint invalid: `reason`                                                                                                                                                                                                    |
| ConfigurationGenerated | False            | MaxPipelinesExceeded            | Maximum pipeline count limit exceeded                                                                                                                                                                                                     |
| ConfigurationGenerated | False            | ReferencedSecretMissing         | One or more referenced Secrets are missing: Secret 'my-secret' of Namespace 'my-namespace'                                                                                                                                                |
| ConfigurationGenerated | False            | ReferencedSecretMissing         | One or more keys in a referenced Secret are missing: Key 'my-key' in Secret 'my-secret' of Namespace 'my-namespace'"                                                                                                                      |
| ConfigurationGenerated | False            | ReferencedSecretMissing         | Secret reference is missing field/s: (field1, field2, ...)                                                                                                                                                                                |
| ConfigurationGenerated | False            | TLSCertificateExpired           | TLS (CA) certificate expired on YYYY-MM-DD                                                                                                                                                                                                |
| ConfigurationGenerated | False            | TLSConfigurationInvalid         | TLS configuration invalid                                                                                                                                                                                                                 |
| ConfigurationGenerated | False            | ValidationFailed                | Pipeline validation failed due to an error from the Kubernetes API server                                                                                                                                                                 |
| TelemetryFlowHealthy   | True             | FlowHealthy                     | No problems detected in the telemetry flow                                                                                                                                                                                                |
| TelemetryFlowHealthy   | False            | GatewayAllTelemetryDataDropped  | Backend is not reachable or rejecting metrics. All metrics are dropped. See troubleshooting: [No Metrics Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=no-metrics-arrive-at-the-backend)          |
| TelemetryFlowHealthy   | False            | GatewayBufferFillingUp          | Buffer nearing capacity. Incoming log rate exceeds export rate. See troubleshooting: [Gateway Buffer Filling Up](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=gateway-buffer-filling-up)                                |
| TelemetryFlowHealthy   | False            | GatewayThrottling               | Metric gateway is unable to receive metrics at current rate. See troubleshooting: [Gateway Throttling](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=gateway-throttling)                                                 |
| TelemetryFlowHealthy   | False            | GatewaySomeTelemetryDataDropped | Backend is reachable, but rejecting metrics. Some metrics are dropped. See troubleshooting: [Not All Metrics Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=not-all-metrics-arrive-at-the-backend) |
| TelemetryFlowHealthy   | False            | ConfigurationNotGenerated       | No metrics delivered to backend because MetricPipeline specification is not applied to the configuration of Metric gateway. Check the 'ConfigurationGenerated' condition for more details                                                 |
| TelemetryFlowHealthy   | Unknown          | GatewayProbingFailed            | Could not determine the health of the telemetry flow because the self monitor probing failed                                                                                                                                              |
