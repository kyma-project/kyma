# Serverless

The `serverlesses.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the Serverless configuration that you want to install on your cluster. To get the up-to-date CRD and show the output in the YAML format, run this command:

   ```bash
   kubectl get crd serverlesses.operator.kyma-project.io -o yaml
   ```

## Sample Custom Resource

The following Serverless custom resource (CR) shows configuration of Serverless with the external registry, custom endpoints for eventing and tracing and custom additional configuration.

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
   metadata:
     finalizers:
     - serverless-operator.kyma-project.io/deletion-hook
     name: default
     namespace: kyma-system
   spec:
     dockerRegistry:
       enableInternal: false
       secretName: my-secret
     eventing:
        endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     tracing:
        endpoint: http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces
     targetCPUUtilizationPercentage: 50
     functionRequeueDuration: 5m
     functionBuildExecutorArgs: "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--use-new-run,--compressed-caching=false"
     functionBuildMaxSimultaneousJobs: 5
     healthzLivenessTimeout: "10s"
     defaultBuildJobPreset: "normal"
     defaultRuntimePodPreset: "M"
     logLevel: "info"
     logFormat: "json"
   status:
     conditions:
     - lastTransitionTime: "2023-04-28T10:09:37Z"
       message: Configured with default Publisher Proxy URL and default Trace Collector
         URL.
       reason: Configured
       status: "True"
       type: Configured
     - lastTransitionTime: "2023-04-28T10:15:15Z"
       message: Serverless installed
       reason: Installed
       status: "True"
       type: Installed
     eventPublisherProxyURL: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     state: Ready
     traceCollectorURL: http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces
   ```

## Custom Resource Parameters

For details, see the [Serverless specification file](https://github.com/kyma-project/serverless-manager/blob/main/components/operator/api/v1alpha1/serverless_types.go).
<!-- TABLE-START -->
### Serverless.operator.kyma-project.io/v1alpha1

**Spec:**

| Parameter                                 | Type    | Description                                                                                                                                          |
|-------------------------------------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| **dockerRegistry**                        | object  | **Deprecated: a future Serverless version won't build images**                                                                                       |
| **dockerRegistry.&#x200b;enableInternal** | boolean | When set to `true`, the internal Docker registry is enabled                                                                                          |
| **dockerRegistry.&#x200b;secretName**     | string  | Secret used for configuration of the Docker registry                                                                                                 |
| **eventing**                              | object  |                                                                                                                                                      |
| **eventing.&#x200b;endpoint** (required)  | string  | Used Eventing endpoint                                                                                                                               |
| **tracing**                               | object  |                                                                                                                                                      |
| **tracing.&#x200b;endpoint** (required)   | string  | Used Tracing endpoint                                                                                                                                |
| **targetCPUUtilizationPercentage**        | string  | **Deprecated: a future Serverless version won't create HPAs** Sets a custom CPU utilization threshold for scaling Function Pods                      |
| **functionRequeueDuration**               | string  | Sets the requeue duration for Function. By default, the Function associated with the default configuration is requeued every 5 minutes               |
| **functionBuildExecutorArgs**             | string  | **Deprecated: a future Serverless version won't build images** Specifies the arguments passed to the Function build executor                         |
| **functionBuildMaxSimultaneousJobs**      | string  | **Deprecated: a future Serverless version won't build images** A number of simultaneous jobs that can run at the same time. The default value is `5` |
| **healthzLivenessTimeout**                | string  | Sets the timeout for the Function health check. The default value in seconds is `10`                                                                 |
| **defaultBuildJobPreset**                 | string  | **Deprecated: a future Serverless version won't build images** Configures the default build Job preset to be used                                    |
| **defaultRuntimePodPreset**               | string  | Configures the default runtime Pod preset to be used                                                                                                 |
| **logLevel**                              | string  | Sets desired log level to be used. The default value is "info"                                                                                       |
| **logFormat**                             | string  | Sets desired log format to be used. The default value is "json"                                                                                      |

**Status:**

| Parameter                                            | Type       | Description                                                                                                                                                                                                                                                                                                                                                    |
|------------------------------------------------------|------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **conditions**                                       | \[\]object | Conditions associated with CustomStatus.                                                                                                                                                                                                                                                                                                                       |
| **conditions.&#x200b;lastTransitionTime** (required) | string     | Specifies the last time the condition transitioned from one status to another. This should be when the underlying condition changes.  If that is not known, then using the time when the API field changed is acceptable.                                                                                                                                      |
| **conditions.&#x200b;message** (required)            | string     | Provides a human-readable message indicating details about the transition. This may be an empty string.                                                                                                                                                                                                                                                        |
| **conditions.&#x200b;observedGeneration**            | integer    | Represents the **.metadata.generation** that the condition was set based upon. For instance, if **.metadata.generation** is currently `12`, but the **.status.conditions[x].observedGeneration** is `9`, the condition is out of date with respect to the current state of the instance.                                                                       |
| **conditions.&#x200b;reason** (required)             | string     | Contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field and whether the values are considered a guaranteed API. The value should be a camelCase string. This field may not be empty.                                        |
| **conditions.&#x200b;status** (required)             | string     | Specifies the status of the condition. The value is either `True`, `False`, or `Unknown`.                                                                                                                                                                                                                                                                      |
| **conditions.&#x200b;type** (required)               | string     | Specifies the condition type in camelCase or in `foo.example.com/CamelCase`. Many **.conditions.type** values are consistent across resources like `Available`, but because arbitrary conditions can be useful (see **.node.status.conditions**), the ability to deconflict is important. The regex it matches is `(dns1123SubdomainFmt/)?(qualifiedNameFmt)`. |
| **dockerRegistry**                                   | string     | Used registry configuration. Contains registry URL or "internal".                                                                                                                                                                                                                                                                                              |
| **eventingEndpoint**                                 | string     | Used Eventing endpoint.                                                                                                                                                                                                                                                                                                                                        |
| **served** (required)                                | string     | Served signifies that current Serverless is managed. Value can be one of `True`, or `False`.                                                                                                                                                                                                                                                                   |
| **state**                                            | string     | Signifies the current state of Serverless. Value can be one of `Ready`, `Processing`, `Error`, or `Deleting`.                                                                                                                                                                                                                                                  |
| **tracingEndpoint**                                  | string     | Used Tracing endpoint.                                                                                                                                                                                                                                                                                                                                         |
| **targetCPUUtilizationPercentage**                   | string     | Used target CPU utilization percentage.                                                                                                                                                                                                                                                                                                                        |
| **functionRequeueDuration**                          | string     | Used the Function requeue duration.                                                                                                                                                                                                                                                                                                                            |
| **functionBuildExecutorArgs**                        | string     | Used the Function build executor arguments.                                                                                                                                                                                                                                                                                                                    |
| **functionBuildMaxSimultaneousJobs**                 | string     | Used the Function build max number of simultaneous jobs.                                                                                                                                                                                                                                                                                                       |
| **healthzLivenessTimeout**                           | string     | Used the healthz liveness timeout.                                                                                                                                                                                                                                                                                                                             |
| **defaultBuildJobPreset**                            | string     | Used the default build Job preset.                                                                                                                                                                                                                                                                                                                             |
| **defaultRuntimePodPreset**                          | string     | Used the default runtime Pod preset.                                                                                                                                                                                                                                                                                                                           |
| **logLevel**                                         | string     | Used the log level.                                                                                                                                                                                                                                                                                                                                            |
| **logFormat**                                        | string     | Used the log format.                                                                                                                                                                                                                                                                                                                                           |

<!-- TABLE-END -->

### Status Reasons

Processing of a Serverless CR can succeed, continue, or fail for one of these reasons:

## Serverless CR Conditions

This section describes the possible states of the Serverless CR. Three condition types, `Installed`, `Configured` and `Deleted`, are used.

| No | CR State   | Condition type | Condition status | Condition reason      | Remark                                        |
|----|------------|----------------|------------------|-----------------------|-----------------------------------------------|
| 1  | Processing | Configured     | true             | Configured            | Serverless configuration verified             |
| 2  | Processing | Configured     | unknown          | ConfigurationCheck    | Serverless configuration verification ongoing |
| 3  | Error      | Configured     | false            | ConfigurationCheckErr | Serverless configuration verification error   |
| 7  | Error      | Configured     | false            | ServerlessDuplicated  | Only one Serverless CR is allowed             |
| 4  | Ready      | Installed      | true             | Installed             | Serverless workloads deployed                 |
| 5  | Processing | Installed      | unknown          | Installation          | Deploying serverless workloads                |
| 6  | Error      | Installed      | false            | InstallationErr       | Deployment error                              |
| 8  | Deleting   | Deleted        | unknown          | Deletion              | Deletion in progress                          |
| 9  | Deleting   | Deleted        | true             | Deleted               | Serverless module deleted                     |
| 10 | Error      | Deleted        | false            | DeletionErr           | Deletion failed                               |
