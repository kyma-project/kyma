# Istio Resources

## Overview

[Istio](https://istio.io/) is an open-source service mesh providing a uniform way to integrate microservices, manage traffic flow across microservices, enforce policies, and aggregate telemetry data.

The Istio Resources Helm chart includes Kyma configuration of Istio and consists of:

- Istio monitoring configuration details providing Grafana dashboards specification
- Istio Ingress Gateway configuring incoming traffic to Kyma
- Mutual TLS (mTLS) configuration enabling mTLS cluster-wide in the STRICT mode
- Service Monitor configuring monitoring for the Istio component
- Istio [Virtual Service](https://istio.io/docs/reference/config/networking/virtual-service/) informing whether Istio is up and running

## Prerequisites

Installation of Istio Resources chart requires Kyma prerequisties, namely [`cluster essentials`](../cluster-essentials),[`istio`](../istio), and [`certificates`](../certificates), to be installed first.

## Installation

To install Istio Resources, run:

```bash
kyma deploy --component istio-resources
```

For more details regarding the installation of Istio itself in Kyma, see the [Istio chart](../istio/README.md).
