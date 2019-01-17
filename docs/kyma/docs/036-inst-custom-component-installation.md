---
title: Custom component installation
type: Installation
---

You can use Kyma with a custom deployment of a component that you installed in the target environment. To enable such implementation, remove a given component from the list of components that install with Kyma.
The version of your component's deployment must match the version that Kyma currently supports. [todo]

## Remove a component

To disable a component from the list of components that install with Kyma, remove this component's entries from the appropriate file. The file depends on whether you install Kyma from the release or from sources, and whether you install Kyma locally or on the cluster.

### Installation from the release

1. Download one of the Kyma [release](https://github.com/kyma-project/kyma/releases).
2. Configure the file:
  * If you want to install Kyma release **locally** without a given component, remove the component's **name** and **namespace** from the `kyma-config-local.yaml` file. For example, if you want to disable Istio installation, remove these lines:
  ```
  name: "istio"
  namespace: "istio-system"
  ```
  * If you want to install Kyma release on a **cluster** without a given component, remove the component's **name** and **namespace** from the `kyma-config-cluster.yaml` file.

3. Follow the installation steps described in the [Install Kyma locally from the release](#installation-install-kyma-locally-from-the-release) document, or [Install Kyma on a GKE cluster](#installation-install-kyma-on-a-gke-cluster) accordingly.

### Installation from sources

1. [todo]
2. Configure the file:
  * If you want to install Kyma from sources **locally** without a given component, remove the component's **name** and **namespace** from the [installer-cr.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) file.

  * If you want to install Kyma from sources on a **cluster** without a given component, remove the component's **name** and **namespace** from the [installer-cr-cluster.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) file.

3. Follow the installation steps described in the [Install Kyma locally from sources](#installation-install-kyma-locally-from-sources) document, or [Install Kyma on a GKE cluster](#installation-install-kyma-on-a-gke-cluster) accordingly.

### Verify the installation

1. Check if all Pods are running in the `kyma-system` Namespace:
  ```
  kubectl get pods -n kyma-system
  ```
2. Sign in to the Kyma Console using the `admin@kyma.cx` as described in the **Install Kyma locally from the release** document.


## Add a component
[todo]
