# Istio Resources

## Overview

[Istio](https://istio.io/) is an open-source service mesh providing a uniform way to integrate microservices, manage traffic flow across microservices, enforce policies, and aggregate telemetry data.

The Istio Resources Helm chart includes Kyma configuration of Istio and consists of:

- Istio monitoring configuration details
- Istio Ingress Gateway
- Mutual TLS (mTLS) configuration enabling mTLS cluster-wide in a STRICT mode
- Service Monitor
- Virtual Service informing whether Istio is up and running

For more details regarding installation of Istio itself in Kyma, see the [Istio chart](../istio-configuration/README.md).
