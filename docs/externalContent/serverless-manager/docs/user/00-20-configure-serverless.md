# Configuring Serverless

By default, the Serverless module comes with the default configuration. You can change the configuration using the Serverless CustomResourceDefinition (CRD), which manages Serverless custom resource (CR).

## Prerequisites

- You have the [Serverless module added](https://kyma-project.io/#/02-get-started/01-quick-install).

- You have access to Kyma dashboard. Alternatively, to use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl).

## Context

The Serverless module has its own operator (Serverless Operator). It watches the Serverless custom resource (CR) and reconfigures (reconciles) the Serverless workloads.

The Serverless CR is an API to configure the Serverless module. You can use it to perform the following actions:

- Enable or disable the internal Docker registry.
- Configure the external Docker registry.
- Override endpoint for traces collected by the Serverless Functions.
- Override endpoint for Eventing.
- Override the target CPU utilization percentage.
- Override the Function requeue duration.
- Override the Function build executor arguments.
- Override the Function build max simultaneous jobs.
- Override the healthz liveness timeout.
- Override the Function request body limit.
- Override the Function timeout.
- Override the default build Job preset.
- Override the default runtime Pod preset.
- Override the default log level.
- Override the default log format.
- Enable network policies.
- Enable buildless mode of Serverless.

The default configuration of the Serverless Module is following:

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
   metadata:
     name: serverless-sample
   spec:
     dockerRegistry:
       enableInternal: true
   ```

> [!WARNING]
> The `spec.dockerRegistry` field is deprecated and will be removed in a future version of Serverless where Functions won't require building images.

## Procedure

1. Go to Kyma dashboard. The URL is in the Overview section of your subaccount.
2. Choose **Modify Modules**, and in the **View** tab, choose `serverless`.
4. Go to **Edit**, and provide your configuration changes. You can use the **Form** or **YAML** tab.

### Configuring Docker Registry

By default, Serverless uses PersistentVolume (PV) as the internal registry to store Docker images for Functions. The default storage size of a single volume is 20 GB. This internal registry is suitable for local development.

> [!WARNING]
> If you use Serverless for production purposes, it is recommended that you use an external registry, such as Docker Hub, Artifact Registry, or Azure Container Registry (ACR).

Follow these steps to use the external Docker registry in Serverless:

1. Create a Secret in the `kyma-system` namespace with the required data (`username`, `password`, `serverAddress`, and `registryAddress`):

   ```bash
   kubectl create secret generic my-registry-config \
       --namespace kyma-system \
       --from-literal=username={USERNAME} \
       --from-literal=password={PASSWORD} \
       --from-literal=serverAddress={SERVER_URL} \
       --from-literal=registryAddress={REGISTRY_URL}
   ```

   > [!TIP]
   > In case of DockerHub, usually the Docker registry address is the same as the account name.

   Examples:

<!-- tabs:start -->

#### Docker Hub

      ```bash
      kubectl create secret generic my-registry-config \
         --namespace kyma-system \
         --from-literal=username={USERNAME} \
         --from-literal=password={PASSWORD} \
         --from-literal=serverAddress=https://index.docker.io/v1/ \
         --from-literal=registryAddress={USERNAME}
      ```

#### Artifact Registry

      ```bash
      kubectl create secret generic my-registry-config \
          --namespace kyma-system \
          --from-literal=username=_json_key \
          --from-literal=password={GCR_KEY_JSON} \
          --from-literal=serverAddress=gcr.io \
          --from-literal=registryAddress=gcr.io/{YOUR_GCR_PROJECT}
      ```

   For more information on how to set up authentication for Docker with Artifact Registry, see the [Artifact Registry documentation](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key).

#### ACR

      ```bash
      kubectl create secret generic my-registry-config \
          --namespace kyma-system \
          --from-literal=username=00000000-0000-0000-0000-000000000000 \
          --from-literal=password={ACR_TOKEN} \
          --from-literal=serverAddress={AZ_REGISTRY_NAME}.azurecr.io \
          --from-literal=registryAddress={AZ_REGISTRY_NAME}.azurecr.io
      ```

   For more information on how to authenticate with ACR, see the [ACR documentation](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-authentication?tabs=azure-cli#az-acr-login-with---expose-token).

<!-- tabs:end -->

2. Reference the Secret in the Serverless CR:

   ```yaml
   spec:
     dockerRegistry:
       secretName: my-registry-config 
   ```

The URL of the currently used Docker registry is visible in the Serverless CR status.

### Configuring Trace Endpoint

By default, the Serverless operator checks if there is a trace endpoint available. If available, the detected trace endpoint is used as the trace collector URL in Functions.
If no trace endpoint is detected, Functions are configured with no trace collector endpoint.
You can configure a custom trace endpoint so that Function traces are sent to any tracing backend you choose.
The currently used trace endpoint is visible in the Serverless CR status.

   ```yaml
   spec:
     tracing:
       endpoint: http://jaeger-collector.observability.svc.cluster.local:4318/v1/traces
   ```

### Configuring Eventing Endpoint

You can configure a custom Eventing endpoint to publish events sent from your Functions.
The currently used trace endpoint is visible in the Serverless CR status.
By default `http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish` is used.

   ```yaml
   spec:
     eventing:
       endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
   ```

### Configuring Target CPU Utilization Percentage

You can set a custom target threshold for CPU utilization. The default value is set to `50%`.

```yaml
   spec:
      targetCPUUtilizationPercentage: 50
```

> [!WARNING]
> The `spec.targetCPUUtilizationPercentage` field is deprecated and will be removed in a future version of Serverless, where automatic HPA creation will be disabled.

### Configuring the Function Requeue Duration

By default, the Function associated with the default configuration will be requeued every 5 minutes.  

```yaml
   spec:
      functionRequeueDuration: 5m
```

### Configuring the Function Build Executor Arguments

Use this label to choose the [arguments](https://github.com/GoogleContainerTools/kaniko?tab=readme-ov-file#additional-flags) passed to the Function build executor, for example:

- `--insecure` - executor operates in an insecure mode
- `--skip-tls-verify` - executor skips the TLS certificate verification
- `--skip-unused-stages` - executor skips any stages that aren't used for the current execution
- `--log-format=text` - executor uses logs in a given format
- `--cache=true` - enables caching for the executor
- `--compressed-caching=false` - prevents tar compression for cached layers. This increases the runtime of the build, but decrease the memory usage especially for large builds.
- `--use-new-run` - improves performance by avoiding the full filesystem snapshots.

```yaml
   spec:
      functionBuildExecutorArgs: "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--use-new-run,--compressed-caching=false"
```

> [!WARNING]
> The `spec.functionBuildExecutorArgs` field is deprecated and will be removed in a future version of Serverless where Functions won't require building images.

### Configuring the Function Build Max Simultaneous Jobs

You can set a custom maximum number of simultaneous jobs which can run at the same time. The default value is set to `5`.

```yaml
   spec:
      functionBuildMaxSimultaneousJobs: 5
```

> [!WARNING]
> The `spec.functionBuildMaxSimultaneousJobs` field is deprecated and will be removed in a future version of Serverless where Functions won't require building images.

### Configuring the healthz Liveness Timeout

By default, Function is considered unhealthy if the liveness health check endpoint does not respond within 10 seconds.

```yaml
   spec:
      healthzLivenessTimeout: "10s"
```

### Configuring the Default Build Job Preset

You can configure the default build Job preset to be used.

```yaml
   spec:
      defaultBuildJobPreset: "normal"
```

> [!WARNING]
> The `spec.defaultBuildJobPreset` field is deprecated and will be removed in a future version of Serverless where Functions won't require building images.

### Configuring the Default Runtime Pod Preset

You can configure the default runtime Pod preset to be used.

```yaml
   spec:
      defaultRuntimePodPreset: "M"
```

For more information on presets, see [Available Presets](https://kyma-project.io/#/serverless-manager/user/technical-reference/07-80-available-presets).

### Configuring the Log Level

You can configure the desired log level to be used.

```yaml
   spec:
      logLevel: "debug"
```

### Configuring the Log Format

You can configure the desired log format to be used.

```yaml
   spec:
      logFormat: "yaml"
```

### Enabling Network Policies

You can enable built-in network policies to ensure that the necessary communication channels required by Serverless workloads remain functional,
even on Kubernetes clusters where strict "deny-all" network policies are enforced. This allows Serverless components to operate correctly
by permitting essential traffic while maintaining a secure cluster environment.

```yaml
   spec:
      enableNetworkPolicies: true
```

### Enabling Buildless Mode

> [!WARNING]
> Buildless mode is a feature flag that you can enable through an annotation. Before enabling the feature, see [Serverless Buildless Mode](technical-reference/03-10-buildless-serverless.md).

You can enable buildless mode of Serverless to skip the image build step for Functions, accelerating prototype development by eliminating the need to build and push custom Function images.

   ```yaml
    annotations:
      serverless.kyma-project.io/buildless-mode: "enabled"
   ```
