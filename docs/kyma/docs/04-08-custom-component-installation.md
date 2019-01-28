---
title: Custom component installation
type: Installation
---

Since Kyma is modular, you can remove some components so that they are not installed together with Kyma. You can also add some of them after the installation process. Read this document to learn how to do that.

## Remove a component

>**NOTE:** Not all components can be simply removed from the Kyma installation. In case of Istio and the Service Catalog, you must provide your own deployment of these components in the Kyma-supported version before you remove them from the installation process. See [this](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/templates/job.yaml#L25) and [this](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/values.yaml#L3) files accordingly to check the currently supported versions of Istio and the Service Catalog.

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

To install a component that does is not installed with Kyma by default, run an appropriate `helm install` command inside the `resources` directory:

* To install Jaeger, run:

```bash
helm install -n jaeger -f jaeger/values.yaml --namespace kyma-system --set global.domainName=kyma.local --set global.isLocalEnv=true jaeger/
```

* To install Logging, run:

```bash
helm install logging --set global.isLocalEnv=true --namespace=kyma-system --name=logging
```

* To install Monitoring, run:

```bash
helm install monitoring --set global.isLocalEnv=true --set global.alertTools.credentials.victorOps.apikey="" --set global.alertTools.credentials.victorOps.routingkey="" --set global.alertTools.credentials.slack.channel="" --set global.alertTools.credentials.slack.apiurl="" --set global.domainName=kyma.local --namespace=kyma-system --name=monitoring
```
