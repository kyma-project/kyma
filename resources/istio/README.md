# Istio

## Overview

[Istio](https://istio.io/) is an open-source service mesh providing a uniform way to integrate microservices, manage traffic flow across microservices, enforce policies, and aggregate telemetry data.

The documentation here is for developers only, please follow the installation instructions from [istio.io](https://istio.io/docs/setup/install/istioctl/) for all other use cases.

The Istio Helm chart consists of the `istio-operator.yaml` file with tailored Istio including Kyma-specific changes and configuration options.

By default, this chart installs the following Istio components:

- ingressgateway
- egressgateway
- istiod (pilot, citadel, and galley)

To disable any of the components, change the corresponding `enabled` flag.

## Installation

The installation of the Istio chart requires [Reconciler](https://github.com/kyma-incubator/reconciler/tree/main/pkg/reconciler/instances/istio). Reconciler uses `istioctl` and a rendered `istio-operator.yaml` file to install Istio on a cluster. To install the component, run:

```bash
kyma deploy --components istio@istio-system
```

## Configuration

The installation of Istio ships with a default configuration. There may be circumstances in which you want to change the defaults.

Istio offers an Istio Control Plane CR, which is used to configure the installation. See the list of the currently exposed parameters of the Istio component that you can override. To learn more, go to [`istio-operator.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/istio/templates/istio-operator.yaml).

- **mesh.Config**
- **values.global**
- **values.pilot**
- **values.sidecarInjectorWebhook**

To override the default values, read how to [change Kyma settings](../../docs/04-operation-guides/operations/03-change-kyma-config-values.md).
