---
title: Custom component installation
type: Configuration
---

By default, you install Kyma with a set of components provided in the [**Kyma Lite**](#installation-overview) package.

During installation, the Kyma Installer applies the content of the local or cluster installation file that includes the list of component names and Namespaces in which the components are installed. The Installer skips the lines starting with a hash character (#):

```
# - name: "backup"
#   namespace: "kyma-system"
```

You can modify the component list as follows:

- Add components to the installation file before the installation
- Add components to the installation file after the installation
- Remove components from the installation file before the installation

>**NOTE:** Currently, it is not possible to remove a component that is already installed. If you remove it from the installation file or precede its entries with a hash character (#) when Kyma is already installed, the Kyma Installer simply does not update this component during the update process but the component is not removed.

Each modification requires an action from the Kyma Installer for the changes to take place:
- If you make changes before the installation, proceed with the standard installation process to finish Kyma setup.
- If you make changes after the installation, follow the [update process](#installation-update-kyma) to refresh the current setup.

Read the subsections for details.

## Add a component

You can add a component before and after installation.

### Installation from the release

1. Download the [newest version](https://github.com/kyma-project/kyma/releases) of Kyma.

2. Customize the installation by adding a component to the list in the installation file or removing the hash character (#) in front of the `name` and `namespace` entries. For example, to enable the Monitoring installation, add or unmark these entries:

    ```
    - name: "monitoring"
      namespace: "kyma-system"
    ```

    * in the `kyma-installer-local.yaml` file for the **local** installation
    * in the `kyma-installer-cluster.yaml` file for the **cluster** installation

3. Follow the installation steps to [install Kyma locally from the release](#installation-install-kyma-locally) or [install Kyma on a cluster](#installation-install-kyma-on-a-cluster).

### Installation from sources

1. Customize the installation by adding a component to the list of components or removing the hash character (#) in front of the `name` and `namespace` entries in the following installation files:

   * [`installer-cr.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) for the **local** installation
   *  [`installer-cr-cluster.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) for the **cluster** installation

2. Follow the installation steps to [install Kyma locally from sources](#installation-install-kyma-locally) or [install Kyma on a cluster](#installation-install-kyma-on-a-cluster).

### Post-installation changes

To add a component that was not installed with Kyma by default, modify the Installation custom resource.

1. Edit the resource:
    ```
    kubectl edit installation kyma-installation
    ```
2. Add the new component to the list of components or remove the hash character (#) preceding these lines:
    ```
    #- name: "jaeger"
    #  namespace: "kyma-system"
    ```
3. Trigger the installation:

   ```
   kubectl -n default label installation/kyma-installation action=install
   ```

### Verify the installation

You can verify the installation status by calling `./installation/scripts/is-installed.sh` in the terminal.

## Remove a component

You can only remove the component before the installation process starts. To disable a component on the list of components that you install with Kyma by default, remove this component's `name` and `namespace` entries from the appropriate installation file or add hash character (#) in front of them. The file differs depending on whether you install Kyma from the release or from sources, and if you install Kyma locally or on a cluster. The version of your component's deployment must match the version that Kyma currently supports.

>**NOTE:** For some components, you must perform additional actions to remove them from the Kyma installation. In case of Istio and the Service Catalog, you must provide your own deployment of these components in the Kyma-supported version before you remove them from the installation process. See [this](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/templates/job.yaml#L25) file to check the currently supported version of Istio. See [this](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/values.yaml#L3) file to check the currently supported version of the Service Catalog.

### Installation from the release

1. Download the [newest version](https://github.com/kyma-project/kyma/releases) of Kyma.
2. Customize the installation by removing a component from the list in the installation file or adding a hash character (#) in front of the `name` and `namespace` entries. For example, to disable the Application Connector installation, remove these entries or add a hash character (#) in front of:
    ```
    - name: "application-connector"
    namespace: "kyma-system"
    ```
  * in the `kyma-installer-local.yaml` file for the **local** installation
  * in the `kyma-installer-cluster.yaml` file for the **cluster** installation

3. Follow the installation steps to [install Kyma locally from the release](#installation-install-kyma-locally) or [install Kyma on a cluster](#installation-install-kyma-on-a-cluster).

### Installation from sources

1. Customize the installation by removing a component from the list of components or adding a hash character (#) in front of the `name` and `namespace` entries in the following installation files:
  * [`installer-cr.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) for the **local** installation
  *  [`installer-cr-cluster.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) for the **cluster** installation

2. Follow the installation steps to [install Kyma locally from sources](#installation-install-kyma-locally) or [install Kyma on a cluster](#installation-install-kyma-on-a-cluster).

### Verify the installation

1. Check if all Pods are running in the `kyma-system` Namespace:
  ```
  kubectl get pods -n kyma-system
  ```
2. Sign in to the Kyma Console using the `admin@kyma.cx` email address as described in the [Install Kyma locally from the release](#installation-install-kyma-locally) document.
