# Istio

## Overview

[Istio](https://istio.io/) is an open-source service mesh providing a uniform way to integrate microservices, manage traffic flow across microservices, enforce policies, and aggregate telemetry data.

The Istio Helm chart includes the `istio-manager-config.yaml` file, which contains the configuration needed for the Istio module's installation.

By default, the Istio module installs the following Istio components:

- ingressgateway
- istiod (pilot, citadel, and galley)
- istio-cni

## Installation

The installation of the Istio chart requires [Reconciler](https://github.com/kyma-incubator/reconciler/tree/main/pkg/reconciler/instances/istio). Reconciler uses the [Istio library](https://github.com/istio/istio/tree/master/operator) and the `IstioOperatorConfiguration` file to install Istio on a cluster. To install the component, run:

```bash
kyma deploy --components istio@istio-system
```

## Configuration

The installation of Istio ships with a default configuration. There may be circumstances in which you want to change the defaults.

Istio is installed as a module, which can be configured using the Istio custom resource. Visit the [Istio repository](https://github.com/kyma-project/istio) for more information.
