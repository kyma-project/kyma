---
title: Istio patch
type: Details
---

As a core component, Istio installs with every Kyma deployment by default. The installation consists of two steps:

1. Istio installs using the official, raw charts from the currently supported release. The charts that are currently
used are stored in the `resources/istio` directory. The installation is customized by enabling security in Istio.

>**NOTE:** Every installation of Istio for Kyma must have security enabled.

2. A custom Istio patch is applied to further customize the Istio installation. A Kubernetes job introduces these changes:
  - Sets a memory limit for every sidecar.
  - Modifies Istio components to use Zipkin in the `kyma-system` Namespace, instead of the default `istio-system`.
  - Adds a webhook to the Istio Pilot.
  - Creates a TLS certificate for the Ingress Gateway.
  - Deletes all resources related to the `prometheus`, `tracing`, `grafana`, and `servicegraph`charts.

To learn more about the custom Istio patch applied in Kyma, see the `components/istio-kyma-patch/` directory.

## Use an existing Istio installation with Kyma

You can use an existing installation of Istio running on Kubernetes with Kyma. The custom Istio patch is applied to such an installation.

>**NOTE:** You cannot skip applying the Istio patch in the Kyma installation process.

To allow such implementation, you must install Kyma without Istio. Read the **Installation with custom Istio deployment** document in the **Kyma**
topic for more details.
