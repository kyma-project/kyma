---
title: Custom component installation
type: Installation
---

Since Kyma is a modular tool, you can remove some components so that they do not install together with Kyma. You can also add some of them after the installation process. Read this document to learn how.

## Remove a component

To disable a component from the list of components that install with Kyma, remove this component's entries from the appropriate file. The file depends on whether you install Kyma from the release or from sources, and whether you install Kyma locally or on a cluster. The version of your component's deployment must match the version that Kyma currently supports.

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

1. Configure the file:
  * If you want to install Kyma from sources **locally** without a given component, remove the component's **name** and **namespace** from the [installer-cr.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) file.

  * If you want to install Kyma from sources on a **cluster** without a given component, remove the component's **name** and **namespace** from the [installer-cr-cluster.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) file.

2. Follow the installation steps described in the [Install Kyma locally from sources](#installation-install-kyma-locally-from-sources) document, or [Install Kyma on a GKE cluster](#installation-install-kyma-on-a-gke-cluster) accordingly.

### Verify the installation

1. Check if all Pods are running in the `kyma-system` Namespace:
  ```
  kubectl get pods -n kyma-system
  ```
2. Sign in to the Kyma Console using the `admin@kyma.cx` as described in the **Install Kyma locally from the release** document.


## Add a component

To install a component that does not install with Kyma by default, run an appropriate `helm install` command inside the `resources` directory:

* To install Jaeger, run:

```bash
helm install -n jaeger -f jaeger/values.yaml --namespace kyma-system --set-string global.domainName=kyma.local --set-string global.isLocalEnv=true jaeger/
```

* To install Logging, run:

```bash
helm install logging --set global.isLocalEnv=true --namespace=kyma-system --name=logging
```

* To install Monitoring, run:

```bash
helm install monitoring --set global.isLocalEnv=true --set global.alertTools.credentials.victorOps.apikey="" --set global.alertTools.credentials.victorOps.routingkey="" --set global.alertTools.credentials.slack.channel="" --set global.alertTools.credentials.slack.apiurl="" --set global.domainName=kyma.local --namespace=kyma-system --name=monitoring
```
