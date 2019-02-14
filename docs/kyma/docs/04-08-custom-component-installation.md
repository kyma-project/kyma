---
title: Custom component installation
type: Installation
---

Since Kyma is modular, you can remove some components so that they are not installed together with Kyma. You can also add some of them after the installation process. Read this document to learn how to do that.

## Remove a component

>**NOTE:** Not all components can be simply removed from the Kyma installation. In case of Istio and the Service Catalog, you must provide your own deployment of these components in the Kyma-supported version before you remove them from the installation process. See [this](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/templates/job.yaml#L25) file to check the currently supported version of Istio. See [this](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/values.yaml#L3) file to check the currently supported version of the Service Catalog.

To disable a component from the list of components that you install with Kyma, remove this component's entries from the appropriate file. The file differs depending on whether you install Kyma from the release or from sources, and if you install Kyma locally or on a cluster. The version of your component's deployment must match the version that Kyma currently supports.

### Installation from the release

1. Download the [newest version](https://github.com/kyma-project/kyma/releases) of Kyma.
2. Customize installation by removing a component from the list of components in the Installation resource. For example, to disable the Application Connector installation, remove this entry:
    ```
    name: "application-connector"
    namespace: "kyma-system"
    ```
  * from the `kyma-config-local.yaml` file for the **local** installation
  * from the `kyma-config-cluster.yaml` file for the **cluster** installation


3. Follow the installation steps described in the [Install Kyma locally from the release](#installation-install-kyma-locally-from-the-release) document, or [Install Kyma on a GKE cluster](#installation-install-kyma-on-a-gke-cluster) accordingly.

### Installation from sources

1. Customize installation by removing a component from the list of components in the following Installation resource:
  * [installer-cr.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) for the **local** installation
  *  [installer-cr-cluster.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) for the **cluster** installation

2. Follow the installation steps described in the [Install Kyma locally from sources](#installation-install-kyma-locally-from-sources) document, or [Install Kyma on a GKE cluster](#installation-install-kyma-on-a-gke-cluster) accordingly.

### Verify the installation

1. Check if all Pods are running in the `kyma-system` Namespace:
  ```
  kubectl get pods -n kyma-system
  ```
2. Sign in to the Kyma Console using the `admin@kyma.cx` email address as described in the [Install Kyma locally from the release](#installation-install-kyma-locally-from-the-release) document.


## Add a component

>**NOTE:** This section assumes that you already have your Kyma Lite local version installed successfully.

To install a component that is not installed with Kyma by default, modify the [Installation](#custom-resource-installation) custom resource and add the component that you want to install to the list of components :

1. Edit the resource:
    ```
    kubectl edit installation kyma-installation
    ```
2. Add the new component to the list of components, for example:
    ```
    - name: "jaeger"
      namespace: "kyma-system"
    ```
3. Trigger the installation:
   ```
   kubectl label installation/kyma-installation action=install
   ```

You can verify the installation status by calling `./installation/scripts/is-installed.sh` in the terminal.
