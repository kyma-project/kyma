---
title: Istio installation
type: Details
---

As a core component, Istio installs with every Kyma deployment by default. The installation consists of two steps:

1. Istio installs using the official, raw charts from the currently supported release. The charts that are currently 
used are stored in the resources/istio directory. The installation is customized by enabling security in Istio.
    >**NOTE:** Every installation of Istio for Kyma must have security enabled.

2. A custom Istio Patch is applied to further customize the Istio installation. A Kubernetes job introduces these 
changes:
  - Sets a memory limit for every sidecar
  - Istio components use Zipkin in the `kyma-system` Namespace, instead of the default `istio-system`
  - A webhook is added to the Istio Pilot
  - A TLS certificate is created for the Ingress Gateway
  - All resources related to the `prometheus`, `tracing`, `grafana`, and `servicegraph`charts  are deleted

During installation raw official charts from currently supported Istio release are installed on cluster. The only 
customization done at this point of installation is enabling security in Istio. This is prerequisite to run Kyma and 
every Istio installation for Kyma needs to have security enabled.

You can find more details about those changes in component `istio-kyma-patch`.

## Use an existing Istio installation with Kyma

You can use an existing installation of Istio running on Kubernetes with Kyma. The custom Istio patch is applied to such 
an installation.

>**NOTE:** You cannot skip applying the Istio patch in the Kyma installation process.

To allow such an implementation, you must install Kyma without Istio. To achieve this, modify the Installation custom 
resource used to trigger the Kyma installer. Read the Installation document in the Custom Resource section of the Kyma 
topic for more details.
