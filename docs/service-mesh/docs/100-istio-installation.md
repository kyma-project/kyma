---
title: Istio installation
type: Details
---

Istio is a foundation Kyma is built upon. It is required for almost every functionality Kyma provides and it needs to 
contain some changes in original charts in order to run Kyma.

## Installation

By default Istio is installed during Kyma installation. The process is split into two parts:
1. Installation of official istio charts. Charts are vendored within Kyma repository under `resources/istio`. 
2. Patch of Istio to make it suitable for Kyma. 

During installation raw official charts from currently supported Istio release are installed on cluster. The only 
customization done at this point of installation is enabling security in Istio. This is prerequisite to run Kyma and 
every Istio installation for Kyma needs to have security enabled.
 
Second stage, patch is done by kubernetes job which introduces following changes into Istio:
* Every sidecar have memory limits set
* Istio components uses Zipkin in `kyma-system` Namespace, instead of `istio-system`
* Webhook is added to pilot 
* TLS certificate for Ingress Gateway are created
* All resources from charts `prometheus`, `tracing`, `grafana` and `servicegraph` are deleted

You can find more details about those changes in component `istio-kyma-patch`.

## Using your own Istio

It is possible to install Kyma with Istio already installed. If you have running kubernetes with istio on it you can 
add Kyma to the cluster.

>**NOTE:** Bear in mind, that patch will always be applied and will change existing Istio installation as described 
above.

To install Kyma without installing istio you need to modify Installation CR. You will find it in the installation 
configuration file which you should create as first step of installation on any cloud provider. Refer to instructions 
specific for your cloud provider to find more details about creation of the 
installation configuration file.

Installation configuration file contains list of components to install in property `spec.components`. One of entries
is named `istio`. To disable Istio installation along with Kyma you need to remove that entry completely. To find more
information about structure of Installation CR visit **Kyma** > **Custom Resource** > **Installation**