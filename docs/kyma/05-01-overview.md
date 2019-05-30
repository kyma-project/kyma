---
title: Overview
type: Configuration
---

## Introduction

Users can configure the way Kyma is installed using in two ways:
  - Customize the list of the components to install
  - Provide overrides that allow to set specific configuration values for a component or entire installation.

The list of components to install is specified in the **Installation** Custom Resource which, once defined in the cluster, is processed by Kyma Installer to perform the installation.
During the installation of a component, Kyma Installer finds and applies all defined overrides in order to customize the configuration used in the installation process.
Detailed description of Overrides is [here](https://kyma.project.io).


## Default settings

Default configuration settings depends on installation type.

### Cluster installation
For cluster installations refer to the relevant cluster installation procedure documented [here](https://kyma.project.io/).
There is a list of components available for every Kyma release in the release artifact named *kyma-installer-cluster.yaml*.
The overrides, if any, are described in relevant cluster installation documentation. The set of overrides is small, since default values are embedded in Helm charts of released Kyma version.

### Local installation
For local installation the list of available components is stored in Kyma project sources in the file **kyma/installation/resources/installer-cr.yaml.tpl**
Default installation overrides can be found in the file **kyma/installation/resources/installer-config-local.yaml.tpl**
>**CAUTION:** The configuration files contain tested and recommended settings. Note that you modify them on your own risk.


## Installation configuration

Before you start the Kyma installation process, you can customize the default settings.

### Components

One of the installation artifacts is the Kyma-Installer image that bundles the Installer executable with all component charts located in the `kyma/resources` folder and corresponding to some Kyma version.
Both cluster and local installation files contain a full list of these component names and their Namespaces.
Components that are not an integral part of the default Kyma Lite package are commented out with a hash character (#). It means that the Installer skips them during installation.
You can customize component installation files by:
- Uncommenting component entries to enable a component installation.
- Adding hashtags to disable a component installation.
- Adding new components to the list along with their chart definition in the `kyma/resources` folder. In that case you must create your own [Kyma Installer image](#installation-use-your-own-kyma-installer-image) as you are adding new component to Kyma.

For more details on custom component installation, see [this](#configuration-custom-component-installation) document.

### Overrides

#### Cluster installations
If you need to customize the installation beyond what's already described in the relevant [cluster installation](https://kyma.project.io) documents, you can define new Overrides.
Remember that Overrides must exist in a cluster before the installation is started, otherwise the Installer is not able to apply them.
Overrides that affect entire installation, e.g. domainName, are already described in cluster installation documentation.
Other overrides are component-specific. To learn the available configuration options for a given component, refer to **Configuration** section of the component documentation.
Once you know the name and possible set of values for an override, you can define it by extending an existing ConfigMap/Secret object, or creating a new one.
[Read more](#configuration-helm-overrides-for-kyma-installation) about the types of overrides and the rules for creating them.

> **CAUTION:** In documentation for each Kyma component, you can find configuration documents that list configurable parameters from the `values.yaml` files for a given component chart or sub-chart. Override values only for the parameters exposed in those configuration documents.

## Runtime configuration

Once Kyma installation completes, you can still modify the installation artifacts. However, for these changes to take effect, you must trigger the update process:

```
kubectl label installation/kyma-installation action=install
```

>**NOTE:** You cannot uninstall a Kyma component that is already installed. If you remove it from any of the installation files or add hashtags in front of its `name` and `namespace` entries, you only disable its further updates.

Apart from modifying the installation artifacts, you can use ConfigMaps and Secrets to configure the installed components and their behavior in the runtime.

For an example of such a runtime configuration, see [Helm Broker configuration](/components/helm-broker/#configuration-configuration) in which you add links to the bundle repositories to a ConfigMap and label it with the `helm-broker-repo=true` label for the Helm Broker to expose additional Service Classes in the Service Catalog.

## Advanced configuration

All `values.yaml` files in charts and sub-charts contain pre-defined attributes that are:
- Configurable
- Used in chart templates
- Recommended production settings

You can only override values that are included in the `values.yaml` files of a given resource. If you want to extend the available customization options, request adding a new attribute with a default value to the pre-defined list in `values.yaml`. Raise a pull request in which you propose changes in the chart, the new attribute, and its value to be added to `values.yaml`. This way you ensure that you can override these values when needed, without these values being overwritten each time an update or rebase takes place.

>**NOTE:** Avoid modifications of such open-source components as Istio or Service Catalog as such changes can severely impact their future version updates.
