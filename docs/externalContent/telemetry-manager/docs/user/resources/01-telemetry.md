# Telemetry

The `telemetry.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define a Telemetry module instance. To get the current CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd telemetry.operator.kyma-project.io -o yaml
```

## Sample Custom Resource

The following Telemetry object defines a module:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: Telemetry
metadata:
  name: default
  namespace: kyma-system
  generation: 2
spec:
  enrichments:
    cluster:
      name: "clusterName"
    extractPodLabels:
      - key: "app.kubernetes.io/name"
      - keyPrefix: "app.kubernetes.io"
  trace:
    gateway:
      scaling:
        type: Static
        static:
          replicas: 3
  metric:
    gateway:
      scaling:
        type: Static
        static:
          replicas: 4
Status:
  state: Ready
  endpoints:
    traces:
      grpc: http://telemetry-otlp-traces.kyma-system:4317
      http: http://telemetry-otlp-traces.kyma-system:4318
    metrics:
      grpc: http://telemetry-otlp-metrics.kyma-system:4317
      http: http://telemetry-otlp-metrics.kyma-system:4318
  conditions:
    - lastTransitionTime: "2023-09-01T15:28:28Z"
      message: All log components are running
      observedGeneration: 2
      reason: ComponentsRunning
      status: "True"
      type: LogComponentsHealthy
    - lastTransitionTime: "2023-09-01T15:46:59Z"
      message: All metric components are running
      observedGeneration: 2
      reason: ComponentsRunning
      status: "True"
      type: MetricComponentsHealthy
    - lastTransitionTime: "2023-09-01T15:35:38Z"
      message: All trace components are running
      observedGeneration: 2
      reason: ComponentsRunning
      status: "True"
      type: TraceComponentsHealthy
```

For further examples, see the [samples](https://github.com/kyma-project/telemetry-manager/tree/main/config/samples) directory.

## Custom Resource Parameters

For details, see the [Telemetry specification file](https://github.com/kyma-project/telemetry-manager/blob/main/apis/operator/v1alpha1/telemetry_types.go).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
### Telemetry.operator.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **enrichments**  | object | Enrichments configures optional enrichments of all telemetry data collected by pipelines. This field is optional. |
| **enrichments.&#x200b;cluster**  | object | Cluster provides user-defined cluster definitions to enrich resource attributes. |
| **enrichments.&#x200b;cluster.&#x200b;name** (required) | string | Name specifies a custom cluster name for the resource attribute `k8s.cluster.name`. |
| **enrichments.&#x200b;extractPodLabels**  | \[\]object | ExtractPodLabels specifies the list of Pod labels to be used for enrichment. This field is optional. |
| **enrichments.&#x200b;extractPodLabels.&#x200b;key**  | string | Key specifies the exact label key to be used. This field is optional. |
| **enrichments.&#x200b;extractPodLabels.&#x200b;keyPrefix**  | string | KeyPrefix specifies a prefix for label keys to be used. This field is optional. |
| **log**  | object | Log configures module settings specific to the log features. This field is optional. |
| **log.&#x200b;gateway**  | object | Gateway configures the log gateway. |
| **log.&#x200b;gateway.&#x200b;scaling**  | object | Scaling defines which strategy is used for scaling the gateway, with detailed configuration options for each strategy type. |
| **log.&#x200b;gateway.&#x200b;scaling.&#x200b;static**  | object | Static is a scaling strategy enabling you to define a custom amount of replicas to be used for the gateway. Present only if Type = StaticScalingStrategyType. |
| **log.&#x200b;gateway.&#x200b;scaling.&#x200b;static.&#x200b;replicas**  | integer | Replicas defines a static number of Pods to run the gateway. Minimum is 1. |
| **log.&#x200b;gateway.&#x200b;scaling.&#x200b;type**  | string | Type of scaling strategy. Default is none, using a fixed amount of replicas. |
| **metric**  | object | Metric configures module settings specific to the metric features. This field is optional. |
| **metric.&#x200b;gateway**  | object | Gateway configures the metric gateway. |
| **metric.&#x200b;gateway.&#x200b;scaling**  | object | Scaling defines which strategy is used for scaling the gateway, with detailed configuration options for each strategy type. |
| **metric.&#x200b;gateway.&#x200b;scaling.&#x200b;static**  | object | Static is a scaling strategy enabling you to define a custom amount of replicas to be used for the gateway. Present only if Type = StaticScalingStrategyType. |
| **metric.&#x200b;gateway.&#x200b;scaling.&#x200b;static.&#x200b;replicas**  | integer | Replicas defines a static number of Pods to run the gateway. Minimum is 1. |
| **metric.&#x200b;gateway.&#x200b;scaling.&#x200b;type**  | string | Type of scaling strategy. Default is none, using a fixed amount of replicas. |
| **trace**  | object | Trace configures module settings specific to the trace features. This field is optional. |
| **trace.&#x200b;gateway**  | object | Gateway configures the trace gateway. |
| **trace.&#x200b;gateway.&#x200b;scaling**  | object | Scaling defines which strategy is used for scaling the gateway, with detailed configuration options for each strategy type. |
| **trace.&#x200b;gateway.&#x200b;scaling.&#x200b;static**  | object | Static is a scaling strategy enabling you to define a custom amount of replicas to be used for the gateway. Present only if Type = StaticScalingStrategyType. |
| **trace.&#x200b;gateway.&#x200b;scaling.&#x200b;static.&#x200b;replicas**  | integer | Replicas defines a static number of Pods to run the gateway. Minimum is 1. |
| **trace.&#x200b;gateway.&#x200b;scaling.&#x200b;type**  | string | Type of scaling strategy. Default is none, using a fixed amount of replicas. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | Conditions contain a set of conditionals to determine the State of Status. If all Conditions are met, State is expected to be in StateReady. |
| **conditions.&#x200b;lastTransitionTime** (required) | string | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. |
| **conditions.&#x200b;message** (required) | string | message is a human readable message indicating details about the transition. This may be an empty string. |
| **conditions.&#x200b;observedGeneration**  | integer | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. |
| **conditions.&#x200b;reason** (required) | string | reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty. |
| **conditions.&#x200b;status** (required) | string | status of the condition, one of True, False, Unknown. |
| **conditions.&#x200b;type** (required) | string | type of condition in CamelCase or in foo.example.com/CamelCase. |
| **endpoints**  | object | Endpoints for log, trace, and metric gateway. |
| **endpoints.&#x200b;logs**  | object | Logs contains the endpoints for log gateway supporting OTLP. |
| **endpoints.&#x200b;logs.&#x200b;grpc**  | string | gRPC endpoint for OTLP. |
| **endpoints.&#x200b;logs.&#x200b;http**  | string | HTTP endpoint for OTLP. |
| **endpoints.&#x200b;metrics**  | object | Metrics contains the endpoints for metric gateway supporting OTLP. |
| **endpoints.&#x200b;metrics.&#x200b;grpc**  | string | gRPC endpoint for OTLP. |
| **endpoints.&#x200b;metrics.&#x200b;http**  | string | HTTP endpoint for OTLP. |
| **endpoints.&#x200b;traces**  | object | Traces contains the endpoints for trace gateway supporting OTLP. |
| **endpoints.&#x200b;traces.&#x200b;grpc**  | string | gRPC endpoint for OTLP. |
| **endpoints.&#x200b;traces.&#x200b;http**  | string | HTTP endpoint for OTLP. |
| **state** (required) | string | State signifies current state of Module CR. Value can be one of these three: "Ready", "Deleting", or "Warning". |

<!-- TABLE-END -->

The `state` attribute of the Telemetry CR is derived from the combined state of all the subcomponents, namely, from the condition types `LogComponentsHealthy`, `TraceComponentsHealthy` and `MetricComponentsHealthy`.

### Log Components State

The state of the log components is determined by the status condition of type `LogComponentsHealthy`:

| Condition Status | Condition Reason             | Condition Message                                                                                                                                                                                                                  |
|------------------|------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| True             | ComponentsRunning            | All log components are running                                                                                                                                                                                                     |
| True             | NoPipelineDeployed           | No pipelines have been deployed                                                                                                                                                                                                    |
| True             | TLSCertificateAboutToExpire  | TLS (CA) certificate is about to expire, configured certificate is valid until YYYY-MM-DD                                                                                                                                          |
| False            | AgentNotReady                | Log agent DaemonSet is not ready                                                                                                                                                                                                   |
| False            | ReferencedSecretMissing      | One or more referenced Secrets are missing: Secret 'my-secret' of Namespace 'my-namespace'                                                                                                                                         |
| False            | ReferencedSecretMissing      | One or more keys in a referenced Secret are missing: Key 'my-key' in Secret 'my-secret' of Namespace 'my-namespace'"                                                                                                               |
| False            | ReferencedSecretMissing      | Secret reference is missing field/s: (field1, field2, ...)                                                                                                                                                                         |
| False            | ResourceBlocksDeletion       | The deletion of the module is blocked. To unblock the deletion, delete the following resources: LogPipelines (resource-1, resource-2,...), LogParsers (resource-1, resource-2,...)                                                 |
| False            | TLSCertificateExpired        | TLS (CA) certificate expired on YYYY-MM-DD                                                                                                                                                                                         |
| False            | TLSConfigurationInvalid      | TLS configuration invalid                                                                                                                                                                                                          |
| False            | ValidationFailed             | Pipeline validation failed due to an error from the Kubernetes API server                                                                                                                                                          |
| False            | AgentAllTelemetryDataDropped | Backend is not reachable or rejecting logs. All logs are dropped. See troubleshooting: [No Logs Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/02-logs?id=no-logs-arrive-at-the-backend)                  |
| False            | AgentBufferFillingUp         | Buffer nearing capacity. Incoming log rate exceeds export rate. See troubleshooting: [Agent Buffer Filling Up](https://kyma-project.io/#/telemetry-manager/user/02-logs?id=agent-buffer-filling-up)                                |
| False            | AgentNoLogsDelivered         | Backend is not reachable or rejecting logs. Logs are buffered and not yet dropped. See troubleshooting: [No Logs Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/02-logs?id=no-logs-arrive-at-the-backend) |
| False            | AgentSomeDataDropped         | Backend is reachable, but rejecting logs. Some logs are dropped. See troubleshooting: [No All Logs Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/02-logs?id=not-all-logs-arrive-at-the-backend)          |

### Trace Components State

The state of the trace components is determined by the status condition of type `TraceComponentsHealthy`:

| Condition Status | Condition Reason               | Condition Message                                                                                                                                                                                                       |
|------------------|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| True             | ComponentsRunning              | All trace components are running                                                                                                                                                                                        |
| True             | NoPipelineDeployed             | No pipelines have been deployed                                                                                                                                                                                         |
| True             | TLSCertificateAboutToExpire    | TLS (CA) certificate is about to expire, configured certificate is valid until YYYY-MM-DD                                                                                                                               |
| False            | GatewayNotReady                | Trace gateway Deployment is not ready                                                                                                                                                                                   |
| False            | MaxPipelinesExceeded           | Maximum pipeline count exceeded                                                                                                                                                                                         |
| False            | ReferencedSecretMissing        | One or more referenced Secrets are missing: Secret 'my-secret' of Namespace 'my-namespace'                                                                                                                              |
| False            | ReferencedSecretMissing        | One or more keys in a referenced Secret are missing: Key 'my-key' in Secret 'my-secret' of Namespace 'my-namespace'"                                                                                                    |
| False            | ResourceBlocksDeletion         | The deletion of the module is blocked. To unblock the deletion, delete the following resources: TracePipelines (resource-1, resource-2,...)                                                                             |
| False            | TLSCertificateExpired          | TLS (CA) certificate expired on YYYY-MM-DD                                                                                                                                                                              |
| False            | TLSConfigurationInvalid        | TLS configuration invalid                                                                                                                                                                                               |
| False            | ValidationFailed               | Pipeline validation failed due to an error from the Kubernetes API server                                                                                                                                               |
| False            | GatewayAllTelemetryDataDropped | Backend is not reachable or rejecting spans. All spans are dropped. See troubleshooting: [No Spans Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/03-traces?id=no-spans-arrive-at-the-backend) |
| False            | GatewayBufferFillingUp         | Buffer nearing capacity. Incoming log rate exceeds export rate. See troubleshooting: [Gateway Buffer Filling Up](https://kyma-project.io/#/telemetry-manager/user/03-traces?id=gateway-buffer-filling-up)               |
| False            | GatewayThrottling              | Trace gateway is unable to receive spans at current rate. See troubleshooting: [Gateway Throttling](https://kyma-project.io/#/telemetry-manager/user/03-traces?id=gateway-throttling)                                   |
| False            | GatewaySomeTelemetryDataDropped             | Backend is reachable, but rejecting spans. Some spans are dropped. [No All Spans Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/03-traces?id=not-all-spans-arrive-at-the-backend)              |

### Metric Components State

The state of the metric components is determined by the status condition of type `MetricComponentsHealthy`:

| Condition Status | Condition Reason                | Condition Message                                                                                                                                                                                                                        |
|------------------|---------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| True             | ComponentsRunning               | All metric components are running                                                                                                                                                                                                        |
| True             | NoPipelineDeployed              | No pipelines have been deployed                                                                                                                                                                                                          |
| True             | TLSCertificateAboutToExpire     | TLS (CA) certificate is about to expire, configured certificate is valid until YYYY-MM-DD                                                                                                                                                |
| False            | AgentNotReady                   | Metric agent DaemonSet is not ready                                                                                                                                                                                                      |
| False            | GatewayNotReady                 | Metric gateway deployment is not ready                                                                                                                                                                                                   |
| False            | MaxPipelinesExceeded            | Maximum pipeline count exceeded                                                                                                                                                                                                          |
| False            | ReferencedSecretMissing         | One or more referenced Secrets are missing: Secret 'my-secret' of Namespace 'my-namespace'                                                                                                                                               |
| False            | ReferencedSecretMissing         | One or more keys in a referenced Secret are missing: Key 'my-key' in Secret 'my-secret' of Namespace 'my-namespace'"                                                                                                                     |
| False            | ResourceBlocksDeletion          | The deletion of the module is blocked. To unblock the deletion, delete the following resources: MetricPipelines (resource-1, resource-2,...)                                                                                             |
| False            | TLSCertificateExpired           | TLS (CA) certificate expired on YYYY-MM-DD                                                                                                                                                                                               |
| False            | TLSConfigurationInvalid         | TLS configuration invalid                                                                                                                                                                                                                |
| False            | ValidationFailed                | Pipeline validation failed due to an error from the Kubernetes API server                                                                                                                                                                |
| False            | GatewayAllTelemetryDataDropped  | Backend is not reachable or rejecting metrics. All metrics are dropped. See troubleshooting: [No Metrics Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=no-metrics-arrive-at-the-backend)         |
| False            | GatewayBufferFillingUp          | Buffer nearing capacity. Incoming log rate exceeds export rate. See troubleshooting: [Gateway Buffer Filling Up](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=gateway-buffer-filling-up)                               |
| False            | GatewayThrottling               | Metric gateway is unable to receive metrics at current rate. See troubleshooting: [Gateway Throttling](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=gateway-throttling)                                                |
| False            | GatewaySomeTelemetryDataDropped | Backend is reachable, but rejecting metrics. Some metrics are dropped. See troubleshooting: [No All Metrics Arrive at the Backend](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=not-all-metrics-arrive-at-the-backend) |

### Telemetry CR State

- 'Ready': Only if all the subcomponent conditions (LogComponentsHealthy, TraceComponentsHealthy, and MetricComponentsHealthy) have a status of `True`.
- 'Warning': If any of these conditions are not `True`.
- 'Deleting': When a Telemetry CR is being deleted.
