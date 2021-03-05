---
title: Custom component installation
type: Configuration
---

During deployment, the Kyma Installer applies the content of the [local](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl#L14) or [cluster](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl#L14) deployment file that includes the list of component names and Namespaces in which the components are deployed. The Installer skips the lines starting with a hash character (#):

```yaml
# - name: "tracing"
#   namespace: "kyma-system"
```

You can modify the component list as follows:

- Add components to the deployment file before the deployment
- Add components to the deployment file after the deployment
- Remove components from the deployment file before the deployment

>**NOTE:** Currently, it is not possible to remove a component that is already deployed. If you remove it from the deployment file or precede its entries with a hash character (#) when Kyma is already deployed, the Kyma CLI simply does not update this component during the update process but the component is not removed.

Each modification requires an action from the Kyma Installer for the changes to take place:

- If you make changes before the deployment, proceed with the standard deployment process to finish Kyma setup.
- If you make changes after the deployment, follow the [update process](#installation-update-kyma) to refresh the current setup.

Read the subsections for details.

## Provide a custom list of components

You can provide a custom list of components to Kyma CLI during the deployment. The version of your component's deployment must match the version that Kyma currently supports.

>**NOTE:** For some components, you must perform additional actions to exclude them from the Kyma deployment. In case of the Service Catalog, you must provide your own deployment of this component in the Kyma-supported version before you remove them from the deployment process. See the [`values.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/values.yaml#L3) file for the currently supported version of the Service Catalog.

### Installation from the release

1. Create a file with the list of components you desire to deploy. You can copy and paste most of the components from the regular [installation file](https://github.com/kyma-project/kyma/blob/master/installation/resources/components.yaml), then modify the list as you like. An example file can look like the following:

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
  - name: "knative-eventing"
    namespace: "knative-eventing"
```

2. Follow the deployment steps to [deploy Kyma locally from the release](#installation-install-kyma-locally) or [deploy Kyma on a cluster](#installation-install-kyma-on-a-cluster). While installing, provide the path to the component list file using the `-c` flag.

### Deployment from sources

1. Customize the deployment by adding a component to the list of components or removing the hash character (#) in front of the `name` and `namespace` entries in the following deployment file:

   * [`components.yaml`](https://github.com/kyma-project/kyma/blob/master/installation/resources/components.yaml)

2. Follow the deployment steps to [deploy Kyma locally from sources](#installation-install-kyma-locally) or [deploy Kyma on a cluster](#installation-install-kyma-on-a-cluster).

### Post-deployment changes

You can only add a new component after the deployment. Removal of the deployed components is not possible. To add a component that was not deployed with Kyma by default, perform the following steps.

1. Download the current [Deployment custom resource](#custom-resource-installation) from the cluster:

    ```bash
    kubectl -n default get installation kyma-installation -o yaml > installation.yaml
    ```

2. Add the new component to the list of components or remove the hash character (#) preceding these lines:

    ```yaml
    #- name: "tracing"
    #  namespace: "kyma-system"
    ```

3. Check which version you're currently running. Run this command:

    ```bash
    kyma version
    ```

4. Trigger the update using the same version and the modified deployment file:

   ```bash
   kyma deploy -s {VERSION} -c {DEPLOYMENT_FILE_PATH}
   ```
