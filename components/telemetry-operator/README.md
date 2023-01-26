# Telemetry Operator

## Overview

To implement [Kyma's strategy](https://github.com/kyma-project/community/blob/main/concepts/observability-strategy/strategy.md) of moving from in-cluster observability backends to a Telemetry component that integrates with external backends, the telemetry operator provides APIs for configurable logging, tracing, and monitoring.

The telemetry operator has been bootstrapped with [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) 3.6.0. Additional APIs can also be [added by Kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api.html).

### Configurable Logging

The logging controllers generate a Fluent Bit configuration from one or more LogPipeline and LogParser custom resources. The controllers ensure that all Fluent Bit Pods run the current configuration by restarting Pods after the configuration has changed. See all [CRD attributes](apis/telemetry/v1alpha1/logpipeline_types.go) and some [examples](config/samples).

For now, creating Fluent Bit Pods is out of scope of the operator. An existing Fluent Bit DaemonSet is expected.

The generated ConfigMap (by default, `telemetry-fluent-bit-sections` in the `kyma-system` namespace) must be mounted to the Fluent Bit Pods and consumed by an `@INCLUDE` statement in an existing [configuration file](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/configuration-file). Fluent Bit parsers, file references, and environment variables are available in an additional ConfigMap or Secret.

See the flags that configure all ConfigMaps, Secret and DaemonSet names in [main.go](main.go).

Further design decisions and test results are documented in [Dynamic Logging Backend Configuration](https://github.com/kyma-project/community/tree/main/concepts/observability-strategy/configurable-logging).

### Configurable Tracing

>**Configurable tracing is still in development and not active with the default Kyma settings.**

The trace controller creates an [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) deployment and related Kubernetes objects from a `TracePipeline` custom resource. The collector is configured to receive traces using the OTLP and OpenCensus protocols, and forwards the received traces to a configurable OTLP backend.

See [Dynamic Trace Backend Configuration](https://github.com/kyma-project/community/tree/main/concepts/observability-strategy/configurable-tracing) for further information.

### Configurable Monitoring

Configurable monitoring is not implemented yet. Future plans are documented in [Dynamic Monitoring Backend Configuration](https://github.com/kyma-project/community/tree/main/concepts/observability-strategy/configurable-monitoring).

## Development

### Prerequisites
- Install [kubebuilder 3.6.0](https://github.com/kubernetes-sigs/kubebuilder), which is the base framework for this controller.
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free YAML files during local development.
- Install [Golang 1.19](https://golang.org/dl/) or newer (for local execution).
- Install [Docker](https://www.docker.com/get-started).
- Install [OpenSSL](https://www.openssl.org/) to generate a webhook certificate for local execution.

### Available Commands

For development, you can use the following commands:

- Run all tests and validation

```bash
make
```

- Regenerate YAML manifests (CRDs and ClusterRole)

```bash
make manifests-local
```

- Copy CRDs and ClusterRole to installation directory

```bash
make copy-manifests-local
```

- Install CRDs to cluster in current kubeconfig context

```bash
make install-crds-local
```

- Uninstall CRDs to cluster in current kubeconfig context

```bash
make uninstall-crds-local
```

- Run the operator locally (uses current kubeconfig context)

```bash
kubectl -n kyma-system scale deployment telemetry-operator --replicas=0 # Scale down in-cluster telemetry-operator
make run-local
```

- Build container image and deploy to cluster in current kubeconfig context. Deploy telemetry chart first, as described before. Then run the following commands to deploy your own operator image.

```bash
make build-image DOCKER_PUSH_DIRECTORY=<my container repo>
make push-image DOCKER_PUSH_DIRECTORY=<my container repo> DOCKER_TAG=latest
kubectl -n kyma-system set image deployment telemetry-operator manager=<my container repo>/telemetry-operator:latest
```

## Troubleshooting

### Enable pausing reconciliations
You must pause reconciliations to be able to debug the pipelines and, for example, try out a different pipeline configuration or a different OTel configuration. To pause reconciliations, create a `telemetry-override-config` in the `kyma-system` Namespace.
Here is an example of such a ConfigMap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: telemetry-override-config
  namespace: kyma-system
data:
  override-config: |
    global:
      logLevel: debug
    tracing:
      paused: true
    logging:
      paused: true
```
The `global`, `tracing`, and `logging` fields are optional.


#### Debugging steps
1. Create an overriding `telemetry-override-config` ConfigMap.
2. Perform debugging operations.
3. Remove the created ConfigMap.
4. To reset the debug actions, perform a restart of the telemetry operator.
   ```bash
   kubectl rollout restart deployment -n kyma-system telemetry-operator
   ```

**Caveats**
If you change the pipeline CR when the reconciliation is paused, these changes will not be applied immediately but in a periodic reconciliation cycle of one hour. To reconcile earlier, restart the Telemetry operator.

### Profiling

The Telemetry operator has pprof-based profiling activated and exposed on port 6060. Use port-forwarding to access the pprof endpoint. You can find additional information in the Go [pprof package documentation](https://pkg.go.dev/net/http/pprof).

