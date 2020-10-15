---
title: Custom component installation
type: Configuration
---

By default, you install Kyma with a set of components provided in the [**Kyma Lite**](#installation-overview) package.

During installation, the Kyma Installer applies the content of the [local](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl#L14) or [cluster](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl#L14) installation file that includes the list of component names and Namespaces in which the components are installed. The Installer skips the lines starting with a hash character (#):

```
# - name: "tracing"
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

## Provide a custom list of components

You can provide a custom list of components to Kyma CLI during the installation. The version of your component's deployment must match the version that Kyma currently supports.

>**NOTE:** For some components, you must perform additional actions to exclude them from the Kyma installation. In case of Istio and the Service Catalog, you must provide your own deployment of these components in the Kyma-supported version before you remove them from the installation process. See the [`job.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/templates/job.yaml#L25) file for the currently supported version of Istio, and the [`values.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/values.yaml#L3) file for the currently supported version of the Service Catalog.

### Installation from the release

1. Create a file with the list of components you desire to install. You can copy and paste most of the components from the regular [installation file](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl#L14), then modify the list as you like. An example file can look like the following:

```yaml
components:
  - name: "cluster-essentials"
    namespace: "kyma-system"
  - name: "testing"
    namespace: "kyma-system"
  - name: "istio"
    namespace: "istio-system"
  - name: "xip-patch"
    namespace: "kyma-installer"
  - name: "istio-kyma-patch"
    namespace: "istio-system"
  - name: "knative-eventing"
    namespace: "knative-eventing"
```

2. Follow the installation steps to [install Kyma locally from the release](#installation-install-kyma-locally) or [install Kyma on a cluster](#installation-install-kyma-on-a-cluster). While installing, provide the path to the component list file using the `-c` flag.

### Installation from sources

1. Customize the installation by adding a component to the list of components or removing the hash character (#) in front of the `name` and `namespace` entries in the following installation files:

   * [`installer-cr.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) for the **local** installation
   *  [`installer-cr-cluster.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) for the **cluster** installation

2. Follow the installation steps to [install Kyma locally from sources](#installation-install-kyma-locally) or [install Kyma on a cluster](#installation-install-kyma-on-a-cluster).

### Post-installation changes

You can only add a new component after the installation. Removal of the installed components is not possible. To add a component that was not installed with Kyma by default, modify the Installation custom resource.

1. Edit the resource:

    ```bash
    kubectl -n default edit installation kyma-installation
    ```

2. Add the new component to the list of components or remove the hash character (#) preceding these lines:

    ```yaml
    #- name: "tracing"
    #  namespace: "kyma-system"
    ```

3. Trigger the installation:

   ```bash
   kubectl -n default label installation/kyma-installation action=install
   ```

### Verify the installation

You can verify the installation status by calling `./installation/scripts/is-installed.sh` in the terminal.
