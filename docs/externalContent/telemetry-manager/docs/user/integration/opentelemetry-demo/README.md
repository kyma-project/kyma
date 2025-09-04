# Integrate OpenTelemetry Demo App

## Overview

| Category| |
| - | - |
| Signal types | traces, metrics |
| Backend type | custom in-cluster, third-party remote |
| OTLP-native | yes |

Learn how to install the OpenTelemetry [demo application](https://github.com/open-telemetry/opentelemetry-demo) in a Kyma cluster using a provided [Helm chart](https://github.com/open-telemetry/opentelemetry-helm-charts/tree/main/charts/opentelemetry-demo). The demo application will be configured to push trace data using OTLP to the collector that's provided by Kyma, so that they are collected together with the related Istio trace data.

![setup](./../assets/otel-demo.drawio.svg)

## Table of Content

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Clean Up](#clean-up)

## Prerequisites

- Kyma as the target deployment environment
- The [Telemetry module](../../README.md) is [added](https://kyma-project.io/#/02-get-started/01-quick-install)
- The [Telemetry module](../../README.md) is configured with pipelines for traces and metrics, for example, by following the [SAP CLoud Logging guide](./../sap-cloud-logging/) or [Prometheus](./../prometheus/) and [Loki](./../loki/)
- [Istio Tracing](../../03-traces.md#2-enable-istio-tracing) is enabled
- [Kubectl version that is within one minor version (older or newer) of `kube-apiserver`](https://kubernetes.io/releases/version-skew-policy/#kubectl)
- Helm 3.x

## Installation

### Preparation

1. Export your namespace as a variable with the following command:

    ```bash
    export K8S_NAMESPACE="otel"
    ```

1. If you haven't created a Namespace yet, do it now:

    ```bash
    kubectl create namespace $K8S_NAMESPACE
    ```

1. To enable Istio injection in your Namespace, set the following label:

    ```bash
    kubectl label namespace $K8S_NAMESPACE istio-injection=enabled
    ```

1. Export the Helm release name that you want to use. The release name must be unique for the chosen Namespace. Be aware that all resources in the cluster will be prefixed with that name. Run the following command:

    ```bash
    export HELM_OTEL_RELEASE="otel-demo"
    ```

1. Update your Helm installation with the required Helm repository:

    ```bash
    helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
    helm repo update
    ```

### Install the Application

Run the Helm upgrade command, which installs the chart if not present yet.

```bash
helm upgrade --install --create-namespace -n $K8S_NAMESPACE $HELM_OTEL_RELEASE open-telemetry/opentelemetry-demo -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/opentelemetry-demo/values.yaml
```

The previous command uses the [values.yaml](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/opentelemetry-demo/values.yaml) provided in this `opentelemetry-demo` folder, which contains customized settings deviating from the default settings. The customizations in the provided `values.yaml` cover the following areas:

- Disable the observability tooling provided with the chart
- Configure Kyma Telemetry instead
- Extend memory limits of the demo apps to avoid crashes caused by memory exhaustion
- Adjust initContainers and services of demo apps to work proper with Istio

Alternatively, you can create your own `values.yaml` file and adjust the command.

### Verify the Application

To verify that the application is running properly, set up port forwarding and call the respective local hosts.

1. Verify the frontend:

   ```bash
   kubectl -n $K8S_NAMESPACE port-forward svc/frontend-proxy 8080
   ```

   ```bash
   open http://localhost:8080
   ````

2. Verify that traces and metrics arrive in your backend. Both traces and metrics are enriched with the typical resource attributes like k8s.namespace.name for easy selection.

3. Enable failures with the feature flag service:

   ```bash
   kubectl -n $K8S_NAMESPACE port-forward svc/frontend-proxy 8080
   ```

   ```bash
   open http://localhost:8080/feature/
   ````

4. Generate load with the load generator:

   ```bash
   kubectl -n $K8S_NAMESPACE port-forward svc/frontend-proxy 8080
   ```

   ```bash
   open http://localhost:8080/loadgen/
   ````

## Clean Up

When you're done, you can remove the example and all its resources from the cluster by calling Helm:

```bash
helm delete -n $K8S_NAMESPACE $HELM_OTEL_RELEASE
```
