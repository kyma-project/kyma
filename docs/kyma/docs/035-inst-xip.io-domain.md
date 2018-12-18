---
title: Install Kyma on a GKE cluster with wildcard DNS
type: Installation
---

If you want to try Kyma in a cluster environment without assigning the cluster to a domain you own, you can use [xip.io](www.xip.io) which provides a wildcard DNS for any IP address. Such scenario requires using a self-signed TLS certificate.

This solution is not suitable for a production environment, but makes for a great playground which allows to get to know the product better.

## Prerequisites

The prerequisites match these listed in the **Install Kyma on a GKE cluster**. However, you don't need to prepare a domain for your cluster as it is replaced by a wildcard DNS provided by [xip.io](www.xip.io).

## Installation

The installation process follows the steps outlined in the **Install Kyma on a GKE cluster** document. Skip the DNS configuration of your Google project and start with the **Prepare the GKE cluster** section of the document.

1. In addition to exporting the desired cluster name as an environment, make sure to export your GCP project name. Run:
    ```
    export PROJECT={YOUR_GCP_PROJECT_NAME}
    ```
2. Use this command to prepare a configuration file that deploys Kyma with [xip.io](www.xip.io) providing a wildcard DNS:
    ```
    cat installation/resources/installer.yaml <(echo -e "\n---") installation/resources/installer-config-cluster.yaml.tpl  <(echo -e "\n---") installation/resources/installer-cr-cluster-xip-io.yaml.tpl | sed -e "s/__DOMAIN__//g" |sed -e "s/__TLS_CERT__//g" | sed -e "s/__TLS_KEY__//g" | sed -e "s/__.*__//g" > xip-installer.yaml
    ```
3. Follow the instructions from the **Deploy Kyma** to install Kyma using the configuration file you prepared.

4. Add the custom Kyma [xip.io](www.xip.io) self-signed certificate to the trusted certificates of your OS. For MacOS run:
  ```
  placeholder
  ```

## Access the cluster

placeholder
