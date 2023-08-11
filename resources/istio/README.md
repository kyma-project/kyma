# Istio

## Overview

[Istio](https://istio.io/) is an open-source service mesh providing a uniform way to integrate microservices, manage traffic flow across microservices, enforce policies, and aggregate telemetry data.

The Istio Helm chart consists of the `istio-manager-config.yaml` file with configuration needed for installation of Istio module.

Istio module by default installs the following Istio components:

- ingressgateway
- istiod (pilot, citadel, and galley)
- istio-cni

## Installation

The installation of the Istio chart requires [Reconciler](https://github.com/kyma-incubator/reconciler/tree/main/pkg/reconciler/instances/istio). Reconciler uses [Istio library] and `IstioOperatorConfiguration` file to install Istio on a cluster. To install the component, run:

```bash
kyma deploy --components istio@istio-system
```

## Configuration

The installation of Istio ships with a default configuration. There may be circumstances in which you want to change the defaults.

Istio is installed as a module, that can be configured with Istio Custom Resource. See https://github.com/kyma-project/istio for more information.
